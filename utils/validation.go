package utils

import (
    "github.com/go-playground/locales/id"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    id_translations "github.com/go-playground/validator/v10/translations/id"
    "log"
)

var (
    Validate   *validator.Validate
    Translator ut.Translator
)

func ValidationTranslationInit() {
    Validate = validator.New()

    idLocale := id.New()
    uni := ut.New(idLocale, idLocale)
    trans, _ := uni.GetTranslator("id")
    Translator = trans

    if err := id_translations.RegisterDefaultTranslations(Validate, Translator); err != nil {
        log.Fatalf("💥 Gagal mendaftarkan translasi: %v", err)
    }
}