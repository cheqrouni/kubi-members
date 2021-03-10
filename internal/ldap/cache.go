package ldap

type Cache struct {
	m map[string]*User
}

func (c *Cache) Add(key string, user *User) {
	c.m[key] = user
}

func (c *Cache) Get(key string) *User {
	return c.m[key]
}
