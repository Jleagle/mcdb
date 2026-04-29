package seeder

type Storage interface {
	SaveIP(ip string) error
	SaveLastIP(ip string) error
	LoadLastIP() string
}
