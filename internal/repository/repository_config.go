package repository

type RepositoryConfig struct {
	t        RepositoryType
	host     string
	port     int
	dbName   string
	user     string
	password string
}

func (c *RepositoryConfig) SetRepositoryType(t RepositoryType) *RepositoryConfig {
	c.t = t
	return c
}

func (c *RepositoryConfig) SetHost(host string) *RepositoryConfig {
	c.host = host
	return c
}

func (c *RepositoryConfig) SetPort(port int) *RepositoryConfig {
	c.port = port
	return c
}

func (c *RepositoryConfig) SetDbName(dbName string) *RepositoryConfig {
	c.dbName = dbName
	return c
}

func (c *RepositoryConfig) SetUser(user string) *RepositoryConfig {
	c.user = user
	return c
}

func (c *RepositoryConfig) SetPassword(password string) *RepositoryConfig {
	c.password = password
	return c
}

func (c *RepositoryConfig) RepositoryType() RepositoryType {
	return c.t
}

func (c *RepositoryConfig) Host() string {
	return c.host
}

func (c *RepositoryConfig) Port() int {
	return c.port
}

func (c *RepositoryConfig) DBName() string {
	return c.dbName
}

func (c *RepositoryConfig) User() string {
	return c.user
}

func (c *RepositoryConfig) Password() string {
	return c.password
}
