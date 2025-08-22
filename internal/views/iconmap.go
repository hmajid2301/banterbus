package views

// IconMap maps old templ icon component names to Huge Icons CSS classes
var IconMap = map[string]string{
	"alien":           "hgi hgi-solid hgi-alien",
	"cancelcircle":    "hgi hgi-solid hgi-cancel-circle",
	"circlecheckmark": "hgi hgi-solid hgi-check-circle",
	"clipboard":       "hgi hgi-solid hgi-clipboard",
	"cross":           "hgi hgi-solid hgi-cancel-01",
	"crown":           "hgi hgi-solid hgi-crown",
	"edittext":        "hgi hgi-solid hgi-edit-02",
	"firstplace":      "hgi hgi-solid hgi-medal-first-place",
	"group":           "hgi hgi-solid hgi-group",
	"info":            "hgi hgi-solid hgi-information-circle",
	"play":            "hgi hgi-solid hgi-play",
	"redo":            "hgi hgi-solid hgi-redo-02",
	"remove":          "hgi hgi-solid hgi-delete-02",
	"secondplace":     "hgi hgi-solid hgi-medal-second-place",
	"squarecheckmark": "hgi hgi-solid hgi-tick-square",
	"squarecross":     "hgi hgi-solid hgi-cancel-square",
	"thirdplace":      "hgi hgi-solid hgi-medal-third-place",
	"user":            "hgi hgi-solid hgi-user",
	"warning":         "hgi hgi-solid hgi-alert-triangle",
}

// GetIconClass returns the CSS class for a given icon name
func GetIconClass(iconName string) string {
	if class, exists := IconMap[iconName]; exists {
		return class
	}
	// Default icon if mapping not found
	return "hgi hgi-solid hgi-question-circle"
}
