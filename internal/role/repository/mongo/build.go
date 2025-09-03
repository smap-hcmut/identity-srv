package mongo

import (
	"github.com/nguyentantai21042004/smap-api/internal/models"
	"github.com/nguyentantai21042004/smap-api/internal/role"
	"go.mongodb.org/mongo-driver/bson"
)

func (repo implRepository) buildRoleModel(sc models.Scope, opt role.CreateOptions) models.Role {
	now := repo.clock()
	role := models.Role{
		ID:        repo.db.NewObjectID(),
		Name:      opt.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return role
}

func (repo implRepository) buildUpdate(opt role.UpdateOptions) (models.Role, bson.M) {
	update := bson.M{
		"updated_at": repo.clock(),
		"name":       opt.Name,
	}

	opt.Model.Name = opt.Name

	return opt.Model, update
}