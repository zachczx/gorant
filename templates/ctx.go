package templates

import "context"

// For type safety
func GetCurrentUser(ctx context.Context) string {
	if currentUser, ok := ctx.Value("currentUser").(string); ok {
		return currentUser
	}
	return ""
}

func GetPreferredName(ctx context.Context) string {
	if preferredName, ok := ctx.Value("preferredName").(string); ok {
		return preferredName
	}
	return ""
}

func GetAvatarPath(ctx context.Context) string {
	if avatarPath, ok := ctx.Value("avatarPath").(string); ok {
		return avatarPath
	}
	return ""
}

func GetSortComments(ctx context.Context) string {
	if sortComments, ok := ctx.Value("sortComments").(string); ok {
		return sortComments
	}
	return ""
}

func GetFilter(ctx context.Context) string {
	if filter, ok := ctx.Value("filter").(string); ok {
		return filter
	}
	return ""
}
