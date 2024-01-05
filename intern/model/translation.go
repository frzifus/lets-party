package model

type Translation struct {
	Greeting       string                     `json:"greeting"`
	WelcomeMessage string                     `json:"welcome_message"`
	GuestForm      TranslationGuestForm       `json:"guest_form"`
	Location       TranslationLocationSection `json:"location"`
	Navigation     TranslationNavigation      `json:"navigation"`
	FlagImgSrc     string                     `json:"flag_img_src"`
}

type TranslationGuestForm struct {
	LabelInputFirstname string   `json:"label_input_firstname"`
	LabelInputLastname  string   `json:"label_input_lastname"`
	LabelSelectAge      string   `json:"label_select_age"`
	LabelSelectDiet     string   `json:"label_select_diet"`
	LabelAgeInput       string   `json:"label_age_input"`
	LabelButtonAddGuest string   `json:"label_button_add_guest"`
	LabelButtonSubmit   string   `json:"label_button_submit"`
	SelectOptionsAge    []string `json:"select_options_age"`
	SelectOptionsDiet   []string `json:"select_options_diet"`
}

type TranslationLocationSection struct {
	Title          string `json:"title"`
	OpenExternally string `json:"openExternally"`
}

type TranslationNavigation struct {
	Guests string `json:"guests"`
	Map    string `json:"map"`
}

type LanguageOption struct {
	Lang       string
	FlagImgSrc string
}
