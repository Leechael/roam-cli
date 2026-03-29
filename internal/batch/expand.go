package batch

import (
	"fmt"
	"strings"

	"github.com/Leechael/roamresearch-skills/internal/client"
)

// BlockFinder resolves a block UID by text under a parent.
// Returns empty string (not error) when no match is found.
type BlockFinder interface {
	FindBlockUnderParent(text string, parentUID string) (string, error)
}

// ExpandActions expands high-level DSL actions into native Roam write actions.
// finder may be nil if no attach-to resolution is needed.
func ExpandActions(actions []map[string]any, finder BlockFinder) ([]map[string]any, error) {
	out := make([]map[string]any, 0, len(actions))
	for i, action := range actions {
		typeName, _ := action["action"].(string)
		path := fmt.Sprintf("actions[%d]", i)
		switch typeName {
		case "create-with-children":
			expanded, err := expandCreateWithChildren(action, path, finder)
			if err != nil {
				return nil, err
			}
			out = append(out, expanded...)
		case "create-block":
			block, _ := action["block"].(map[string]any)
			hasChildren := block != nil && block["children"] != nil
			loc, _ := action["location"].(map[string]any)
			hasAttachTo := loc != nil && strings.TrimSpace(stringField(loc, "attach-to")) != ""
			if hasChildren || hasAttachTo {
				expanded, err := expandCreateWithChildren(action, path, finder)
				if err != nil {
					return nil, err
				}
				out = append(out, expanded...)
			} else {
				out = append(out, action)
			}
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

func expandCreateWithChildren(action map[string]any, path string, finder BlockFinder) ([]map[string]any, error) {
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

	// Handle attach-to: find-or-create an intermediate parent
	attachTo := strings.TrimSpace(stringField(loc, "attach-to"))
	if attachTo != "" {
		if finder == nil {
			return nil, fmt.Errorf("%s.location.attach-to requires a connected client", path)
		}
		foundUID, err := finder.FindBlockUnderParent(attachTo, parentUID)
		if err != nil {
			return nil, fmt.Errorf("%s: attach-to lookup failed: %w", path, err)
		}
		if foundUID != "" {
			parentUID = foundUID
		} else {
			// Create the attach-to block, then use it as parent
			newUID := client.NewUID()
			var expanded []map[string]any
			expanded = append(expanded, client.CreateBlockAction(attachTo, parentUID, newUID, rootOrder, true))
			parentUID = newUID
			// Children go under the newly created attach-to block
			block, ok := action["block"].(map[string]any)
			if !ok {
				return expanded, nil
			}
			root, err := parseCreateNode(block, path+".block")
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(root.Order) == "" {
				root.Order = "last"
			}
			appendNodeActions(&expanded, parentUID, root)
			return expanded, nil
		}
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
		uid = client.NewUID()
	}
	order := strings.TrimSpace(node.Order)
	if order == "" {
		order = "last"
	}
	open := true
	if node.Open != nil {
		open = *node.Open
	}
	*actions = append(*actions, client.CreateBlockAction(node.Text, parentUID, uid, order, open))
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
