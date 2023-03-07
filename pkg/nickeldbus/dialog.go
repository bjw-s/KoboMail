// Package nickeldbus implements all NickelDbus interactions of KoboMail
package nickeldbus

import "os/exec"

func dialog(APItype string, args ...string) (stdout string, err error) {
	arg1 := "-m"
	arg2 := APItype
	callArgs := append([]string{arg1, arg2}, args...)
	cmd := exec.Command(binQndb, callArgs...)
	output, err := cmd.Output()

	if err != nil {
		return "", err
	}
	return string(output), nil
}

// DialogCreate creates a dialog to show a notification to the user
func DialogCreate(initialMsg string) {
	if UseNickelDbus {
		dialog("dlgConfirmCreate")
		dialog("dlgConfirmSetTitle", "KoboMail")
		dialog("dlgConfirmSetBody", initialMsg)
		dialog("dlgConfirmSetModal", "false")
		dialog("dlgConfirmShowClose", "true")
		dialog("dlgConfirmShow")
	}
}

// DialogUpdate updates a dialog with a new body
func DialogUpdate(body string) {
	if UseNickelDbus {
		dialog("dlgConfirmSetBody", body)
	}
}

// DialogAddOKButton updates a dialog with a confirmation button
func DialogAddOKButton() {
	if UseNickelDbus {
		dialog("dlgConfirmSetAccept", "OK")
	}
}
