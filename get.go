package ogrnOnline

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	typeQueryCompany = iota
	typeQueryPeople
	typeQueryBusinessman

	host = `https://огрн.онлайн`

	pauseForRequest = 400
)

// isValidQuery проверяет параметры запроса на основе его типа
func isValidQuery(query url.Values, typeQuery int) error {
	testType := map[int]map[string]bool{
		typeQueryCompany:     map[string]bool{"огрн": true, "инн": true, "кпп": true, "наименование": true, "стр": true},
		typeQueryPeople:      map[string]bool{"фамилия": true, "имя": true, "отчество": true, "инн": true, "стр": true},
		typeQueryBusinessman: map[string]bool{"человек": true, "огрнип": true, "инн": true},
	}

	ogrnBusinessman := regexp.MustCompile(`^[0-9]{15}$`)
	ogrnCompany := regexp.MustCompile(`^[0-9]{13}$`)
	innBusinessman := regexp.MustCompile(`^[0-9]{12}$`)
	innCompany := regexp.MustCompile(`^[0-9]{10}$`)

	if typeQuery != typeQueryCompany && typeQuery != typeQueryPeople && typeQuery != typeQueryBusinessman {
		panic(fmt.Errorf("isValidQuery: неверный параметр typeQuery: %d", typeQuery))
	}

	for options := range query {
		if !testType[typeQuery][options] {
			return fmt.Errorf(`неверный параметр "%s"`, options)
		}

		switch {
		case options == "огрн" && typeQuery == typeQueryCompany:
			if !ogrnCompany.MatchString(query[options][0]) {
				return fmt.Errorf("недопустимое значение ОГРН: %s", query[options][0])
			}
		case options == "инн" && typeQuery == typeQueryCompany:
			if !innCompany.MatchString(query[options][0]) {
				return fmt.Errorf("недопустимое значение ИНН: %s", query[options][0])
			}
		case options == "огрнип" && typeQuery == typeQueryBusinessman:
			if !ogrnBusinessman.MatchString(query[options][0]) {
				return fmt.Errorf("недопустимое значение ОГРНИП: %s", query[options][0])
			}
		case options == "инн" && (typeQuery == typeQueryBusinessman || typeQuery == typeQueryPeople):
			if !innBusinessman.MatchString(query[options][0]) {
				return fmt.Errorf("недопустимое значение ИНН: %s", query[options][0])
			}
		}
	}
	return nil
}

// createURL формирует URL на основе пути и запроса
func createURL(path string, query url.Values) *url.URL {
	ur, err := url.Parse(host)
	if err != nil {
		panic(fmt.Errorf("ошибка парсинга хоста: %v", err))
	}
	ur.Path = path
	if query != nil {
		ur.RawQuery = query.Encode()
	}
	return ur
}

// getDataFromServer - базовый запрос, получающий данные от сервера на основе url
func getDataFromServer(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		panic(fmt.Errorf("ошибка запроса к серверу %v", err))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("ошибка чтения ответа сервера: %v", err))
	}
	time.Sleep(time.Millisecond * pauseForRequest)
	return body
}

// FindCompany осуществляет поиск юридического лица по заданным параметрам
func FindCompany(query url.Values) ([]CompanyBaseInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyBaseInfo{}
	)
	if err := isValidQuery(query, typeQueryCompany); err != nil {
		return result, err
	}
	if err := json.Unmarshal(getDataFromServer(createURL(path, query).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// FindPeople осуществляет поиск юридического лица по заданным параметрам
func FindPeople(query url.Values) ([]PeopleInfo, error) {
	var (
		path   = `/интеграция/люди/`
		result = []PeopleInfo{}
	)
	if err := isValidQuery(query, typeQueryPeople); err != nil {
		return result, err
	}
	if err := json.Unmarshal(getDataFromServer(createURL(path, query).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// FindBusinessman осуществляет поиск юридического лица по заданным параметрам
func FindBusinessman(query url.Values) ([]PeopleBusinessmanInfo, error) {
	var (
		path   = `/интеграция/ип/`
		result = []PeopleBusinessmanInfo{}
	)
	if err := isValidQuery(query, typeQueryBusinessman); err != nil {
		return result, err
	}
	if err := json.Unmarshal(getDataFromServer(createURL(path, query).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetCompany возвращает полную информацию о юридическом лице на основе его id
func GetCompany(id int) (CompanyInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = CompanyInfo{}
		param  = "/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetCompany - метод CompanyBaseInfo, возвращаяющий полную информацию о юридическом лице
func (c *CompanyBaseInfo) GetCompany() (CompanyInfo, error) {
	return GetCompany(c.ID)
}

// GetOwners возвращает список участников юридического лица на основе его id
func GetOwners(id int) ([]CompanyOwnerInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyOwnerInfo{}
		param  = "/учредители/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetOwners - метод CompanyBaseInfo, возвращаяющий список участников юридического лица
func (c *CompanyBaseInfo) GetOwners() ([]CompanyOwnerInfo, error) {
	return GetOwners(c.ID)
}

// GetAssociates возвращает список управляющих юридического лица на основе его id
func GetAssociates(id int) ([]CompanyAssociateInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyAssociateInfo{}
		param  = "/сотрудники/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetAssociates - метод CompanyBaseInfo, возвращаяющий список управляющих юридического лица
func (c *CompanyBaseInfo) GetAssociates() ([]CompanyAssociateInfo, error) {
	return GetAssociates(c.ID)
}

// GetSubCompanies возвращает список зависимых компаний юридического лица на основе его id
func GetSubCompanies(id int) ([]CompanyBaseInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyBaseInfo{}
		param  = "/зависимые/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetSubCompanies - метод CompanyBaseInfo, возвращаяющий список зависимых компаний юридического лица
func (c *CompanyBaseInfo) GetSubCompanies() ([]CompanyBaseInfo, error) {
	return GetSubCompanies(c.ID)
}

// GetRepresentativeOffices возвращает список представительсв юридического лица на основе его id
func GetRepresentativeOffices(id int) ([]CompanyBranchInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyBranchInfo{}
		param  = "/представительства/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetRepresentativeOffices - метод CompanyBaseInfo, возвращаяющий список представительсв юридического лица
func (c *CompanyBaseInfo) GetRepresentativeOffices() ([]CompanyBranchInfo, error) {
	return GetRepresentativeOffices(c.ID)
}

// GetBranches возвращает список филиалов юридического лица на основе его id
func GetBranches(id int) ([]CompanyBranchInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyBranchInfo{}
		param  = "/филиалы/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetBranches - метод CompanyBaseInfo, возвращаяющий список филиалов юридического лица
func (c *CompanyBaseInfo) GetBranches() ([]CompanyBranchInfo, error) {
	return GetBranches(c.ID)
}

// GenFinance возвращает бухгалтерские балансы юридиеского лица за предшествующие годы по его id
func GenFinance(id int) ([]CompanyFinanceInfo, error) {
	var (
		path   = `/интеграция/компании/`
		result = []CompanyFinanceInfo{}
		param  = "/финансы/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetFinance - метод объекта CompanyBaseInfo, возвращающий бухгалтерские балансы за предшествующие годы
func (c *CompanyBaseInfo) GetFinance() ([]CompanyFinanceInfo, error) {
	return GenFinance(c.ID)
}

// GetPeople полную информацию о физическом лице на основе его id
func GetPeople(id int) (PeopleInfo, error) {
	var (
		path   = `/интеграция/люди/`
		result = PeopleInfo{}
		param  = "/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetJobs возвращает места работы физического лица на основе его id
func GetJobs(id int) ([]CompanyAssociateInfo, error) {
	var (
		path   = `/интеграция/люди/`
		result = []CompanyAssociateInfo{}
		param  = "/должности/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetJobs - метод PeopleInfo, возвращающий места работы физического лица
func (p *PeopleInfo) GetJobs() ([]CompanyAssociateInfo, error) {
	return GetJobs(p.ID)
}

// GetShare возвращает список компаний c участием физического лица на основе его id
func GetShare(id int) ([]CompanyBaseInfo, error) {
	var (
		path   = `/интеграция/люди/`
		result = []CompanyBaseInfo{}
		param  = "/компании/"
	)
	if err := json.Unmarshal(getDataFromServer(createURL(fmt.Sprintf(`%s%d%s`, path, id, param), nil).String()), &result); err != nil {
		return result, err
	}
	return result, nil
}

// GetShare - метод PeopleInfo, возвращающий список компаний c участием физического лица
func (p *PeopleInfo) GetShare() ([]CompanyBaseInfo, error) {
	return GetShare(p.ID)
}
