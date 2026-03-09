package battle

import (
	"mygame/world/entity"
)

// EnemyTemplate описывает боевые параметры типа врага (registry шаблонов).
type EnemyTemplate struct {
	Name       string
	HP         int
	Attack     int
	Defense    int
	Initiative int
	IsRanged   bool
	Role       Role
}

// GetEnemyTemplate возвращает шаблон врага по kind.
func GetEnemyTemplate(kind entity.EnemyKind) EnemyTemplate {
	switch kind {
	case entity.EnemyKindSlime:
		return EnemyTemplate{
			Name:       "Слайм",
			HP:         6,
			Attack:     1,
			Defense:    0,
			Initiative: 1,
			IsRanged:   false,
			Role:       RoleFighter,
		}
	case entity.EnemyKindWolf:
		return EnemyTemplate{
			Name:       "Волк",
			HP:         8,
			Attack:     2,
			Defense:    1,
			Initiative: 2,
			IsRanged:   true,
			Role:       RoleArcher,
		}
	case entity.EnemyKindBandit:
		return EnemyTemplate{
			Name:       "Бандит",
			HP:         10,
			Attack:     2,
			Defense:    1,
			Initiative: 2,
			IsRanged:   false,
			Role:       RoleHealer,
		}
	default:
		return EnemyTemplate{
			Name:       "Враг",
			HP:         6,
			Attack:     1,
			Defense:    0,
			Initiative: 1,
			IsRanged:   false,
			Role:       RoleFighter,
		}
	}
}
