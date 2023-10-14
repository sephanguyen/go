package unleash

type ManagerOption func(manager *manager)

func NumberOfWorker(numberOfWorker int) ManagerOption {
	return func(manager *manager) {
		manager.numberOfWorker = numberOfWorker
	}
}
