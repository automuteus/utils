package discord

import "testing"

func TestExtractRoleIDFromText(t *testing.T) {
	_, err := ExtractRoleIDFromText("invalid")
	if err == nil {
		t.Error("Expected error for invalid role ID string")
	}

	_, err = ExtractRoleIDFromText("<@&123>")
	if err == nil {
		t.Error("Expected error for invalid role ID string")
	}

	_, err = ExtractRoleIDFromText("<@&141101495071408128")
	if err == nil {
		t.Error("Expected error for invalid role ID string")
	}

	_, err = ExtractRoleIDFromText("<@141101495071408128>")
	if err == nil {
		t.Error("Expected error for invalid role ID string")
	}

	_, err = ExtractRoleIDFromText("<@&-141101495071408128>")
	if err == nil {
		t.Error("Expected error for invalid role ID string")
	}

	id, err := ExtractRoleIDFromText("<@&141101495071408128>")
	if err != nil {
		t.Error("Expected nil error from valid Role ID string <@&141101495071408128>")
	}
	if id != "141101495071408128" {
		t.Error("ID was not extracted correctly")
	}
}

func TestExtractUserIDFromMention(t *testing.T) {
	_, err := ExtractUserIDFromMention("invalid")
	if err == nil {
		t.Error("Expected error for invalid user ID string")
	}

	_, err = ExtractUserIDFromMention("<@123>")
	if err == nil {
		t.Error("Expected error for invalid user ID string")
	}

	_, err = ExtractUserIDFromMention("<@141101495071408128")
	if err == nil {
		t.Error("Expected error for invalid user ID string")
	}

	_, err = ExtractUserIDFromMention("<@-141101495071408128>")
	if err == nil {
		t.Error("Expected error for invalid user ID string")
	}

	id, err := ExtractUserIDFromMention("<@141101495071408128>")
	if err != nil {
		t.Error("Expected nil error from valid Role ID string <@141101495071408128>")
	}
	if id != "141101495071408128" {
		t.Error("ID was not extracted correctly")
	}

	id, err = ExtractUserIDFromMention("<@!141101495071408128>")
	if err != nil {
		t.Error("Expected nil error from valid Role ID string <@!141101495071408128>")
	}
	if id != "141101495071408128" {
		t.Error("ID was not extracted correctly")
	}
}

func TestExtractChannelIDFromMention(t *testing.T) {
	_, err := ExtractChannelIDFromMention("invalid")
	if err == nil {
		t.Error("Expected error for invalid channel ID string")
	}

	_, err = ExtractChannelIDFromMention("<#123>")
	if err == nil {
		t.Error("Expected error for invalid channel ID string")
	}

	_, err = ExtractChannelIDFromMention("<#141101495071408128")
	if err == nil {
		t.Error("Expected error for invalid channel ID string")
	}

	_, err = ExtractChannelIDFromMention("<#-141101495071408128>")
	if err == nil {
		t.Error("Expected error for invalid channel ID string")
	}

	id, err := ExtractChannelIDFromMention("<#141101495071408128>")
	if err != nil {
		t.Error("Expected nil error from valid Channel ID string <#141101495071408128>")
	}
	if id != "141101495071408128" {
		t.Error("ID was not extracted correctly")
	}
}
