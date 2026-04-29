package seeder

type Storage interface {
	SaveIP(ip string) (bool, error)
	SaveLastIP(ip string) error
	LoadLastIP() string
}
