package model

import "github.com/go-webauthn/webauthn/webauthn"

type PasskeyUser struct {
	ID          []byte
	DisplayName string
	Name        string

	Credientials []webauthn.Credential
}

func (u *PasskeyUser) WebAuthnID() []byte {
	return u.ID
}

func (u *PasskeyUser) WebAuthnName() string {
	return u.Name
}

func (u *PasskeyUser) WebAuthnDisplayName() string {
	return u.DisplayName
}

func (u *PasskeyUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credientials
}

func (u *PasskeyUser) WebAuthnIcon() string {
	return ""
}

func (u *PasskeyUser) PutCredential(credential webauthn.Credential) {
	u.Credientials = append(u.Credientials, credential)
}

func (u *PasskeyUser) AddCrediential(credential *webauthn.Credential) {
	u.Credientials = append(u.Credientials, *credential)
}

func (u *PasskeyUser) UpdateCredential(credential *webauthn.Credential) {
	for i, c := range u.Credientials {
		if string(c.ID) == string(credential.ID) {
			u.Credientials[i] = *credential
		}
	}
}
