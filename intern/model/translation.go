// Copyright (C) 2024 the lets-party maintainers
// See root-dir/LICENSE for more information

package model

type Translation struct {
	Title          string                     `json:"title" form:"title"`
	Greeting       string                     `json:"greeting" form:"greeting"`
	WelcomeMessage string                     `json:"welcome_message" form:"welcome_message"`
	FinalMessage   string                     `json:"final_message" form:"final_message"`
	GuestForm      TranslationGuestForm       `json:"guest_form" form:"guest_form"`
	Location       TranslationLocationSection `json:"location" form:"location"`
	Hotels         TranslationHotelsSection   `json:"hotels" form:"hotels"`
	Airports       TranslationAirportsSection `json:"airports" form:"airports"`
	Navigation     TranslationNavigation      `json:"navigation" form:"navigation"`
	FlagImgSrc     string                     `json:"flag_img_src" form:"flag_img_src"`
	Error          Error                      `json:"error" form:"error"`
	Success        Success                    `json:"success" form:"success"`
	And            string                     `json:"and" form:"and"`
}

type TranslationGuestForm struct {
	LabelInputFirstname    string   `json:"label_input_firstname" form:"label_input_firstname"`
	LabelInputLastname     string   `json:"label_input_lastname" form:"label_input_lastname"`
	LabelSelectAge         string   `json:"label_select_age" form:"label_select_age"`
	LabelSelectDiet        string   `json:"label_select_diet" form:"label_select_diet"`
	LabelSelectInvStatus   string   `json:"label_select_inv_status" form:"label_select_inv_status"`
	LabelAgeInput          string   `json:"label_age_input" form:"label_age_input"`
	LabelButtonAddGuest    string   `json:"label_button_add_guest" form:"label_button_add_guest"`
	LabelButtonSubmit      string   `json:"label_button_submit" form:"label_button_submit"`
	SelectOptionsAge       []string `json:"select_options_age" form:"select_options_age"`
	SelectOptionsDiet      []string `json:"select_options_diet" form:"select_options_diet"`
	SelectOptionsInvStatus []string `json:"select_options_inv_status" form:"select_options_inv_status"`
	MessageSubmitSuccess   string   `json:"message_submit_success" form:"message_submit_success"`
}

type TranslationLocationSection struct {
	Title          string `json:"title" form:"title"`
	OpenExternally string `json:"openExternally" form:"openExternally"`
}

type TranslationHotelsSection struct {
	Title   string `json:"title" form:"title"`
	Website string `json:"website" form:"website"`
}

type TranslationAirportsSection struct {
	Title string `json:"title" form:"title"`
}

type TranslationNavigation struct {
	Guests   string `json:"guests" form:"guests"`
	Map      string `json:"map" form:"map"`
	Hotels   string `json:"hotels" form:"hotels"`
	Airports string `json:"airports" form:"airports"`
}

type LanguageOption struct {
	Lang       string `json:"lang" form:"lang"`
	FlagImgSrc string `json:"flagImgSrc" form:"flagImgSrc"`
}

type Error struct {
	Title   string `json:"title" form:"title"`
	Process string `json:"process" form:"process"`
}

type Success struct {
	Title string `json:"title" form:"title"`
}
