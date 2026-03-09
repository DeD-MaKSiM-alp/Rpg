package entity

// AddEntity добавляет сущность в карту entities и возвращает её.
func AddEntity(entities map[EntityID]*Entity, nextID *EntityID, entityType EntityType, kind EnemyKind, worldX, worldY int) *Entity {
	*nextID++
	e := &Entity{
		ID:    *nextID,
		Type:  entityType,
		Kind:  kind,
		X:     worldX,
		Y:     worldY,
		Alive: true,
	}
	entities[e.ID] = e
	return e
}
