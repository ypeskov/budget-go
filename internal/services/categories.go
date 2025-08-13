package services

import (
    "fmt"
    "sort"
    "strings"
    "ypeskov/budget-go/internal/models"
    "ypeskov/budget-go/internal/repositories/categories"
)

type CategoriesService interface {
	GetUserCategories(userId int) ([]models.UserCategory, error)
	GetUserCategoriesGrouped(userId int) (map[string][]models.GroupedCategory, error)
	CreateCategory(name string, isIncome bool, parentID *int, userID int) (*models.UserCategory, error)
	ValidateCategoryOwnership(categoryId int, userId int) (bool, error)
}

type CategoryServiceInstance struct {
	categoriesRepo categories.Repository
}

func NewCategoriesService(repository categories.Repository) CategoriesService {
	return &CategoryServiceInstance{
		categoriesRepo: repository,
	}
}

func (c *CategoryServiceInstance) GetUserCategories(userId int) ([]models.UserCategory, error) {
    userCategories, err := c.categoriesRepo.GetUserCategories(userId)
    if err != nil {
        return nil, err
    }

    // Group by parent so parents come first and children immediately after
    parents := make([]models.UserCategory, 0)
    childrenByParentID := make(map[int][]models.UserCategory)

    for _, category := range userCategories {
        if category.ParentID == nil {
            parents = append(parents, category)
            continue
        }
        parentID := *category.ParentID
        childrenByParentID[parentID] = append(childrenByParentID[parentID], category)
    }

    // Sort parents by name (case-insensitive)
    sort.Slice(parents, func(i, j int) bool {
        return strings.ToLower(parents[i].Name) < strings.ToLower(parents[j].Name)
    })

    ordered := make([]models.UserCategory, 0, len(userCategories))

    for _, parent := range parents {
        // Top-level category: keep name as-is
        ordered = append(ordered, parent)

        // Add its children right after, sorted by name and prefixed with parent name
        if children, ok := childrenByParentID[parent.ID]; ok {
            sort.Slice(children, func(i, j int) bool {
                return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
            })
            for _, child := range children {
                child.Name = fmt.Sprintf("%s >> %s", parent.Name, child.Name)
                ordered = append(ordered, child)
            }
            delete(childrenByParentID, parent.ID)
        }
    }

    // Any remaining children whose parent is missing (orphans): place after, grouped by parentName
    if len(childrenByParentID) > 0 {
        remaining := make([]models.UserCategory, 0)
        for _, list := range childrenByParentID {
            remaining = append(remaining, list...)
        }
        sort.Slice(remaining, func(i, j int) bool {
            var ai, aj string
            if remaining[i].ParentName != nil {
                ai = strings.ToLower(*remaining[i].ParentName)
            }
            if remaining[j].ParentName != nil {
                aj = strings.ToLower(*remaining[j].ParentName)
            }
            if ai == aj {
                return strings.ToLower(remaining[i].Name) < strings.ToLower(remaining[j].Name)
            }
            return ai < aj
        })
        for _, child := range remaining {
            parentName := ""
            if child.ParentName != nil {
                parentName = *child.ParentName
            }
            child.Name = fmt.Sprintf("%s >> %s", parentName, child.Name)
            ordered = append(ordered, child)
        }
    }

    return ordered, nil
}

func (c *CategoryServiceInstance) GetUserCategoriesGrouped(userId int) (map[string][]models.GroupedCategory, error) {
	userCategories, err := c.categoriesRepo.GetUserCategories(userId)
	if err != nil {
		return nil, err
	}

	// Separate categories by type and organize by parent-child relationships
	incomeParents := make([]models.GroupedCategory, 0)
	expenseParents := make([]models.GroupedCategory, 0)
	childrenByParentID := make(map[int][]models.GroupedCategory)

	// First, collect all categories and separate them
	for _, category := range userCategories {
		groupedCategory := models.GroupedCategory{
			ID:       category.ID,
			Name:     category.Name,
			ParentID: category.ParentID,
			IsIncome: category.IsIncome,
			Children: make([]models.GroupedCategory, 0),
		}

		if category.ParentID == nil {
			// This is a parent category
			if category.IsIncome {
				incomeParents = append(incomeParents, groupedCategory)
			} else {
				expenseParents = append(expenseParents, groupedCategory)
			}
		} else {
			// This is a child category
			parentID := *category.ParentID
			childrenByParentID[parentID] = append(childrenByParentID[parentID], groupedCategory)
		}
	}

	// Sort parents alphabetically
	sort.Slice(incomeParents, func(i, j int) bool {
		return strings.ToLower(incomeParents[i].Name) < strings.ToLower(incomeParents[j].Name)
	})
	sort.Slice(expenseParents, func(i, j int) bool {
		return strings.ToLower(expenseParents[i].Name) < strings.ToLower(expenseParents[j].Name)
	})

	// Attach children to their parents
	for i := range incomeParents {
		if children, exists := childrenByParentID[incomeParents[i].ID]; exists {
			sort.Slice(children, func(x, y int) bool {
				return strings.ToLower(children[x].Name) < strings.ToLower(children[y].Name)
			})
			incomeParents[i].Children = children
		}
	}

	for i := range expenseParents {
		if children, exists := childrenByParentID[expenseParents[i].ID]; exists {
			sort.Slice(children, func(x, y int) bool {
				return strings.ToLower(children[x].Name) < strings.ToLower(children[y].Name)
			})
			expenseParents[i].Children = children
		}
	}

	return map[string][]models.GroupedCategory{
		"income":   incomeParents,
		"expenses": expenseParents,
	}, nil
}

func (c *CategoryServiceInstance) CreateCategory(name string, isIncome bool, parentID *int, userID int) (*models.UserCategory, error) {
	// Validate parent category if provided
	if parentID != nil {
		isValidParent, err := c.categoriesRepo.ValidateCategoryOwnership(*parentID, userID)
		if err != nil {
			return nil, err
		}
		if !isValidParent {
			return nil, fmt.Errorf("parent category not found or does not belong to user")
		}
	}

	category := models.UserCategory{
		Name:      name,
		ParentID:  parentID,
		IsIncome:  isIncome,
		UserID:    userID,
		IsDeleted: false,
	}

	return c.categoriesRepo.CreateCategory(category)
}

func (c *CategoryServiceInstance) ValidateCategoryOwnership(categoryId int, userId int) (bool, error) {
	return c.categoriesRepo.ValidateCategoryOwnership(categoryId, userId)
}
