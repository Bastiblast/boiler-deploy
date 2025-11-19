package ansible

// TagCategory représente une catégorie de tags
type TagCategory struct {
	Name        string
	Description string
	Tags        []Tag
}

// Tag représente un tag Ansible avec sa description
type Tag struct {
	Name        string
	Description string
	Selected    bool
}

// GetProvisionTags retourne les catégories de tags pour provision
func GetProvisionTags() []TagCategory {
	return []TagCategory{
		{
			Name:        "System Base",
			Description: "Packages and system configuration",
			Tags: []Tag{
				{Name: "common", Description: "All common tasks", Selected: true},
				{Name: "packages", Description: "Package installation and updates", Selected: true},
				{Name: "apt", Description: "APT operations", Selected: true},
				{Name: "upgrade", Description: "System upgrade", Selected: false},
				{Name: "users", Description: "User management", Selected: true},
				{Name: "config", Description: "System configuration", Selected: true},
			},
		},
		{
			Name:        "Security",
			Description: "Firewall, SSH, and security hardening",
			Tags: []Tag{
				{Name: "security", Description: "All security tasks", Selected: true},
				{Name: "firewall", Description: "Firewall configuration", Selected: true},
				{Name: "ufw", Description: "UFW firewall", Selected: true},
				{Name: "fail2ban", Description: "Fail2ban setup", Selected: true},
				{Name: "ssh", Description: "SSH configuration", Selected: true},
				{Name: "hardening", Description: "Security hardening", Selected: true},
			},
		},
		{
			Name:        "Runtime & Services",
			Description: "Application runtime and web services",
			Tags: []Tag{
				{Name: "nodejs", Description: "Node.js installation", Selected: true},
				{Name: "nginx", Description: "Nginx web server", Selected: true},
				{Name: "postgresql", Description: "PostgreSQL database", Selected: true},
			},
		},
		{
			Name:        "Monitoring",
			Description: "Monitoring and observability",
			Tags: []Tag{
				{Name: "monitoring", Description: "Monitoring tools", Selected: false},
			},
		},
	}
}

// GetDeployTags retourne les catégories de tags pour deploy
func GetDeployTags() []TagCategory {
	return []TagCategory{
		{
			Name:        "Application",
			Description: "Application deployment",
			Tags: []Tag{
				{Name: "deploy", Description: "All deployment tasks", Selected: true},
				{Name: "code", Description: "Code deployment", Selected: true},
				{Name: "health", Description: "Health checks", Selected: true},
			},
		},
	}
}

// FormatTagsForAnsible convertit les tags sélectionnés en string pour Ansible
func FormatTagsForAnsible(categories []TagCategory) string {
	var selectedTags []string
	for _, category := range categories {
		for _, tag := range category.Tags {
			if tag.Selected {
				selectedTags = append(selectedTags, tag.Name)
			}
		}
	}
	
	if len(selectedTags) == 0 {
		return ""
	}
	
	result := ""
	for i, tag := range selectedTags {
		if i > 0 {
			result += ","
		}
		result += tag
	}
	return result
}

// GetAllTags retourne tous les tags sélectionnés
func GetAllTags(categories []TagCategory) []string {
	var tags []string
	for _, category := range categories {
		for _, tag := range category.Tags {
			if tag.Selected {
				tags = append(tags, tag.Name)
			}
		}
	}
	return tags
}
