package domain

import "fmt"

// Query is everything routing may look at: the category and the building.
// There is deliberately no description field (decision 1, FR-11).
type Query struct {
	Category Category
	Building string
}

// Match picks the rule for q. A rule for the exact building beats a rule for
// the whole category (building ""). Among equally specific rules the lowest
// priority number wins. Inactive rules and other categories are ignored.
func Match(rules []Rule, q Query) (Rule, error) {
	if !q.Category.Valid() {
		return Rule{}, fmt.Errorf("%w: %q", ErrInvalidCategory, q.Category)
	}

	var best Rule
	found := false
	for _, r := range rules {
		if !r.Active || r.Category != q.Category {
			continue
		}
		if r.Building != "" && r.Building != q.Building {
			continue
		}
		if !found || better(r, best) {
			best, found = r, true
		}
	}
	if !found {
		return Rule{}, fmt.Errorf("%w: category %s, building %q", ErrRuleNotFound, q.Category, q.Building)
	}
	return best, nil
}

// better reports whether a should be chosen over b. Both already match the query.
func better(a, b Rule) bool {
	aSpecific, bSpecific := a.Building != "", b.Building != ""
	if aSpecific != bSpecific {
		return aSpecific
	}
	return a.Priority < b.Priority
}
