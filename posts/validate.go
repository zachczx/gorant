package posts

import "github.com/rezakhademix/govalidator/v2"

func Validate(p Post) map[string](string) {
	v := govalidator.New()

	v.RequiredString(p.Name, "name", "").RequiredString(p.Content, "content", "").MaxString(p.Content, 2000, "content", "")

	if v.IsFailed() {
		return v.Errors()
	}

	return nil
}
