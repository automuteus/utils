package discord

import (
	"errors"
	"strings"
)

func ExtractRoleIDFromText(mention string) (string, error) {
	// role is formatted <@&123456>
	if strings.HasPrefix(mention, "<@&") && strings.HasSuffix(mention, ">") {
		err := ValidateSnowflake(mention[3 : len(mention)-1])
		if err == nil {
			return mention[3 : len(mention)-1], nil
		}
		return "", err
	} else {
		// if they just used the ID of the role directly
		err := ValidateSnowflake(mention)
		if err == nil {
			return mention, nil
		}
	}
	return "", errors.New("role text does not conform to the correct format (`<@&roleid>` or `roleid`)")
}

func ExtractUserIDFromMention(mention string) (string, error) {
	// nickname format
	switch {
	case strings.HasPrefix(mention, "<@!") && strings.HasSuffix(mention, ">"):
		err := ValidateSnowflake(mention[3 : len(mention)-1])
		if err == nil {
			return mention[3 : len(mention)-1], nil
		}
		return "", err
	case strings.HasPrefix(mention, "<@") && strings.HasSuffix(mention, ">"):
		err := ValidateSnowflake(mention[2 : len(mention)-1])
		if err == nil {
			return mention[2 : len(mention)-1], nil
		}
		return "", err
	default:
		err := ValidateSnowflake(mention)
		if err == nil {
			return mention, nil
		}
		return "", errors.New("mention does not conform to the correct format")
	}
}
