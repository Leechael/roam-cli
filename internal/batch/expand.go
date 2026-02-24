package batch

import (
	"fmt"
	"strings"

	"roam-cli/internal/roam"
)

// ExpandActions expands high-level DSL actions into native Roam write actions.
// Unknown actions are passed through unchanged for backward compatibility.
func ExpandActions(actions []map[string]any) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(actions))
	for i, action := range actions {
		typeName, _ := action["action"].(string)
		switch typeName {
		case "create-with-children":
			expanded, err := expandCreateWithChildren(action, fmt.Sprintf("actions[%d]", i))
			if err != nil {
				return nil, err
			}
			out = append(out, expanded...)
		default:
			out = append(out, action)
		}
	}
	return out, nil
}

type createNode struct {
	UID      string
	Text     string
	Open     *bool
	Order    string
	Children []createNode
}

func expandCreateWithChildren(action map[string]any, path string) ([]map[string]any, error) {
	loc, ok := action["location"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s.location must be an object", path)
	}
	parentUID, _ := loc["parent-uid"].(string)
	if strings.TrimSpace(parentUID) == "" {
		return nil, fmt.Errorf("%s.location.parent-uid is required", path)
	}
	rootOrder, _ := loc["order"].(string)
	if strings.TrimSpace(rootOrder) == "" {
		rootOrder = "last"
	}

	block, ok := action["block"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s.block must be an object", path)
	}
	root, err := parseCreateNode(block, path+".block")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(root.Order) == "" {
		root.Order = rootOrder
	}

	var expanded []map[string]any
	appendNodeActions(&expanded, parentUID, root)
	return expanded, nil
}

func appendNodeActions(actions *[]map[string]any, parentUID string, node createNode) {
	uid := strings.TrimSpace(node.UID)
	if uid == "" {
		uid = roam.NewUID()
	}
	order := strings.TrimSpace(node.Order)
	if order == "" {
		order = "last"
	}
	open := true
	if node.Open != nil {
		open = *node.Open
	}
	*actions = append(*actions, roam.CreateBlockAction(node.Text, parentUID, uid, order, open))
	for _, child := range node.Children {
		appendNodeActions(actions, uid, child)
	}
}

func parseCreateNode(raw map[string]any, path string) (createNode, error) {
	text := firstText(raw)
	if strings.TrimSpace(text) == "" {
		return createNode{}, fmt.Errorf("%s must provide non-empty text via \"string\" or \"text\"", path)
	}

	node := createNode{
		UID:   stringField(raw, "uid"),
		Text:  text,
		Order: stringField(raw, "order"),
	}
	if open, ok := boolField(raw, "open"); ok {
		node.Open = &open
	}

	if childrenRaw, exists := raw["children"]; exists && childrenRaw != nil {
		childList, ok := childrenRaw.([]any)
		if !ok {
			return createNode{}, fmt.Errorf("%s.children must be an array", path)
		}
		node.Children = make([]createNode, 0, len(childList))
		for i, childAny := range childList {
			childMap, ok := childAny.(map[string]any)
			if !ok {
				return createNode{}, fmt.Errorf("%s.children[%d] must be an object", path, i)
			}
			child, err := parseCreateNode(childMap, fmt.Sprintf("%s.children[%d]", path, i))
			if err != nil {
				return createNode{}, err
			}
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}

func firstText(m map[string]any) string {
	if s := stringField(m, "string"); strings.TrimSpace(s) != "" {
		return s
	}
	return stringField(m, "text")
}

func stringField(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func boolField(m map[string]any, key string) (bool, bool) {
	v, ok := m[key].(bool)
	return v, ok
}
