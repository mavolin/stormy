// Package memory provides an in-memory repository.
package memory

type Repository struct {
	*ideaChannelSettingsRepo
	*ideaRepo
}

func New() *Repository {
	return &Repository{
		ideaChannelSettingsRepo: newIdeaChannelSettingsRepo(),
		ideaRepo:                newIdeaRepo(),
	}
}
