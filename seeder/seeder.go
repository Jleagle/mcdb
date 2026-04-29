package seeder

type Storage interface {
	SaveIP(ip string) error
}
