package model

type Account struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Server   string `db:"server" json:"server"`
}

type AccountInfo struct {
	ID             int  `db:"id" json:"id"`
	AccountID      int  `db:"account_id" json:"accountId"`
	Tribe          int  `db:"tribe" json:"tribe"`
	Gold           int  `db:"gold" json:"gold"`
	Silver         int  `db:"silver" json:"silver"`
	HasPlusAccount bool `db:"has_plus_account" json:"hasPlusAccount"`
}

type Access struct {
	ID            int    `db:"id" json:"id"`
	AccountID     int    `db:"account_id" json:"accountId"`
	Username      string `db:"username" json:"username"`
	Password      string `db:"password" json:"password"`
	ProxyHost     string `db:"proxy_host" json:"proxyHost"`
	ProxyPort     int    `db:"proxy_port" json:"proxyPort"`
	ProxyUsername string `db:"proxy_username" json:"proxyUsername"`
	ProxyPassword string `db:"proxy_password" json:"proxyPassword"`
	Useragent     string `db:"useragent" json:"useragent"`
	LastUsed      SQLiteTime `db:"last_used" json:"lastUsed"`
}

// Proxy returns the formatted proxy string or empty if no host.
func (a Access) Proxy() string {
	if a.ProxyHost == "" {
		return ""
	}
	if a.ProxyUsername != "" {
		return a.ProxyUsername + ":" + a.ProxyPassword + "@" + a.ProxyHost
	}
	return a.ProxyHost
}

type AccountSettingRow struct {
	ID        int `db:"id" json:"id"`
	AccountID int `db:"account_id" json:"accountId"`
	Setting   int `db:"setting" json:"setting"`
	Value     int `db:"value" json:"value"`
}
