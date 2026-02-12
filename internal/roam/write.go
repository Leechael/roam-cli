package roam

import "github.com/google/uuid"

func NewUID() string {
	return uuid.NewString()
}

func CreatePageAction(title, uid string) map[string]any {
	if uid == "" {
		uid = NewUID()
	}
	return map[string]any{
		"action": "create-page",
		"page": map[string]any{
			"title":              title,
			"uid":                uid,
			"children-view-type": "bullet",
		},
	}
}

func CreateBlockAction(text, parentUID, uid, order string, open bool) map[string]any {
	if uid == "" {
		uid = NewUID()
	}
	if order == "" {
		order = "last"
	}
	return map[string]any{
		"action": "create-block",
		"location": map[string]any{
			"parent-uid": parentUID,
			"order":      order,
		},
		"block": map[string]any{
			"uid":    uid,
			"string": text,
			"open":   open,
		},
	}
}

func UpdateBlockAction(uid, text string) map[string]any {
	return map[string]any{
		"action": "update-block",
		"block": map[string]any{
			"uid":    uid,
			"string": text,
		},
	}
}

func DeleteBlockAction(uid string) map[string]any {
	return map[string]any{
		"action": "delete-block",
		"block": map[string]any{
			"uid": uid,
		},
	}
}
