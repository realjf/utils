package utils

import (
	"os/user"
	"strconv"
	"sync"
)

type UserAccount struct {
	user     *user.User
	username string
	uid_ui32 uint32
	gid_ui32 uint32
	uid      int
	gid      int
	lock     sync.RWMutex
}

func NewUserAccount() *UserAccount {
	return &UserAccount{
		lock:     sync.RWMutex{},
		uid_ui32: 0,
		gid_ui32: 0,
		uid:      0,
		gid:      0,
		user:     nil,
		username: "",
	}
}

func NewFromUsername(username string) (*UserAccount, error) {
	user, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}
	ua := &UserAccount{
		username: username,
		lock:     sync.RWMutex{},
		uid_ui32: 0,
		gid_ui32: 0,
		uid:      0,
		gid:      0,
		user:     nil,
	}

	ua.SetUser(user)

	return ua, nil
}

func NewFromUser(user *user.User) *UserAccount {
	ua := &UserAccount{
		user:     user,
		lock:     sync.RWMutex{},
		uid_ui32: 0,
		gid_ui32: 0,
		uid:      0,
		gid:      0,
		username: "",
	}
	ua.SetUser(user)
	return ua
}

func (c *UserAccount) SetUser(u *user.User) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.user = u
	c.uid, _ = strconv.Atoi(u.Uid)
	c.gid, _ = strconv.Atoi(u.Gid)
	u64, _ := strconv.ParseUint(u.Uid, 10, 32)
	g64, _ := strconv.ParseUint(u.Gid, 10, 32)
	c.uid_ui32 = uint32(u64)
	c.gid_ui32 = uint32(g64)
	c.username = u.Username
}

func (c *UserAccount) SetUsername(username string) error {
	User, err := user.Lookup(username)
	if err != nil {
		return err
	}
	c.SetUser(User)
	return nil
}

func (c *UserAccount) GetUser() *user.User {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.user
}

func (c *UserAccount) SetUid(u uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.uid = int(u)
	c.uid_ui32 = uint32(u)
}

func (c *UserAccount) GetUid() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.uid
}

func (c *UserAccount) GetUid_ui32() uint32 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.uid_ui32
}

func (c *UserAccount) SetGid(g uint64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.gid = int(g)
	c.gid_ui32 = uint32(g)
}

func (c *UserAccount) GetGid() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.gid
}

func (c *UserAccount) GetGid_ui32() uint32 {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.gid_ui32
}
