package util

import "sync"

// Define a global variable to hold the singleton instance of ApplicationContext.
var instance *applicationContext = nil

// Use a sync.Once variable to ensure that initialization of the instance is done only once.
var once sync.Once

// Define the ApplicationContext interface, which specifies the methods that must be implemented by the singleton instance.
type ApplicationContext interface {
	SetAccessToken(t string)
	GetAcessToken() string
}

// Define the struct that implements the ApplicationContext interface.
type applicationContext struct {
	accessToken  string
	sync.RWMutex // Use sync.RWMutex to handle concurrent access to the accessToken field.
}

// SetAccessToken sets the access token in the singleton instance.
func (a *applicationContext) SetAccessToken(token string) {
	a.Lock()         // Acquire a write lock to ensure exclusive access while setting the access token.
	defer a.Unlock() // Release the write lock when the function completes.
	a.accessToken = token
}

// GetAccessToken retrieves the access token from the singleton instance.
func (a *applicationContext) GetAcessToken() string {
	a.RLock()         // Release the write lock when the function completes.
	defer a.RUnlock() // Release the read lock when the function completes.
	return a.accessToken
}

// GetAccessToken retrieves the access token from the singleton instance.
func GetApplicationContextInstance() ApplicationContext {
	// Use sync.Once to ensure that the instance is created only once, even in the presence of concurrent calls.
	once.Do(func() {
		instance = new(applicationContext) // Create a new instance of the applicationContext.
	})
	return instance
}
