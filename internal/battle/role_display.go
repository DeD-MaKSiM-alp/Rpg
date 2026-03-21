package battle

import "unicode"

// RoleAbbrev — одна буква для маркера роли на токене / в UI (без финального арта).
func RoleAbbrev(r Role) rune {
	switch r {
	case RoleFighter:
		return 'Б'
	case RoleArcher:
		return 'С'
	case RoleHealer:
		return 'Ц'
	case RoleMage:
		return 'М'
	default:
		return '?'
	}
}

// EnemyTokenGlyph — первая значимая буква имени/архетипа для маркера врага.
func EnemyTokenGlyph(u *CombatUnit) string {
	if u == nil {
		return "?"
	}
	s := u.Def.ArchetypeID
	if s == "" {
		s = u.Def.DisplayName
	}
	for _, r := range s {
		if unicode.IsLetter(r) {
			return string(unicode.ToUpper(r))
		}
	}
	return string(rune('0' + (int(u.ID) % 10)))
}
