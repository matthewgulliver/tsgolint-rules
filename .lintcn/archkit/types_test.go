package archkit

import "testing"

func TestNilInputsReturnNothingRatherThanPanic(t *testing.T) {
	t.Parallel()

	if got := Constituents(nil); got != nil {
		t.Errorf("Constituents(nil) = %v, want nil", got)
	}
	if got := DeclaringFiles(nil); got != nil {
		t.Errorf("DeclaringFiles(nil) = %v, want nil", got)
	}
	if got := DeclaringFilesOfSymbol(nil); got != nil {
		t.Errorf("DeclaringFilesOfSymbol(nil) = %v, want nil", got)
	}
}
