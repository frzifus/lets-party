package model

type Translation struct {
	Title          string                     `json:"title"`
	Greeting       string                     `json:"greeting"`
	WelcomeMessage string                     `json:"welcome_message"`
	FinalMessage   string                     `json:"final_message"`
	GuestForm      TranslationGuestForm       `json:"guest_form"`
	Location       TranslationLocationSection `json:"location"`
	Hotels         TranslationHotelsSection   `json:"hotels"`
	Airports       TranslationAirportsSection `json:"airports"`
	Navigation     TranslationNavigation      `json:"navigation"`
	FlagImgSrc     string                     `json:"flag_img_src"`
}

type TranslationGuestForm struct {
	LabelInputFirstname    string   `json:"label_input_firstname"`
	LabelInputLastname     string   `json:"label_input_lastname"`
	LabelSelectAge         string   `json:"label_select_age"`
	LabelSelectDiet        string   `json:"label_select_diet"`
	LabelSelectInvStatus   string   `json:"label_select_inv_status"`
	LabelAgeInput          string   `json:"label_age_input"`
	LabelButtonAddGuest    string   `json:"label_button_add_guest"`
	LabelButtonSubmit      string   `json:"label_button_submit"`
	SelectOptionsAge       []string `json:"select_options_age"`
	SelectOptionsDiet      []string `json:"select_options_diet"`
	SelectOptionsInvStatus []string `json:"select_options_inv_status"`
}

type TranslationLocationSection struct {
	Title          string `json:"title"`
	OpenExternally string `json:"openExternally"`
}

type TranslationHotelsSection struct {
	Title   string `json:"title"`
	Website string `json:"website"`
}

type TranslationAirportsSection struct {
	Title string `json:"title"`
}

type TranslationNavigation struct {
	Guests   string `json:"guests"`
	Map      string `json:"map"`
	Hotels   string `json:"hotels"`
	Airports string `json:"airports"`
}

type LanguageOption struct {
	Lang       string
	FlagImgSrc string
}
