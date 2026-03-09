package battle

import (
	"mygame/world/entity"
)

// EnemyTemplate описывает боевые параметры типа врага.
type EnemyTemplate struct {
	ID         string
	Name       string
	MaxHP      int
	Attack     int
	Defense    int
	Initiative int
}

// GetEnemyTemplate возвращает шаблон врага по kind.
func GetEnemyTemplate(kind entity.EnemyKind) EnemyTemplate {
	switch kind {
	case entity.EnemyKindSlime:
		return EnemyTemplate{
			ID:         "slime",
			Name:       "Слайм",
			MaxHP:      6,
			Attack:     1,
			Defense:    0,
			Initiative: 1,
		}
	case entity.EnemyKindWolf:
		return EnemyTemplate{
			ID:         "wolf",
			Name:       "Волк",
			MaxHP:      8,
			Attack:     2,
			Defense:    1,
			Initiative: 2,
		}
	case entity.EnemyKindBandit:
		return EnemyTemplate{
			ID:         "bandit",
			Name:       "Бандит",
			MaxHP:      10,
			Attack:     2,
			Defense:    1,
			Initiative: 2,
		}
	default:
		return EnemyTemplate{
			ID:         "unknown",
			Name:       "Враг",
			MaxHP:      6,
			Attack:     1,
			Defense:    0,
			Initiative: 1,
		}
	}
}
