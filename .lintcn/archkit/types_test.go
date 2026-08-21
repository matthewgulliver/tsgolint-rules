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
	if got := ElementTypes(nil, nil); got != nil {
		t.Errorf("ElementTypes(nil, nil) = %v, want nil", got)
	}
	if got := Unwrapped(nil, nil); got != nil {
		t.Errorf("Unwrapped(nil, nil) = %v, want nil", got)
	}
	if got := CallSignatures(nil, nil); got != nil {
		t.Errorf("CallSignatures(nil, nil) = %v, want nil", got)
	}
	if IsCallable(nil, nil) {
		t.Errorf("IsCallable(nil, nil) = true, want false")
	}
	if got := Members(nil, nil); got != nil {
		t.Errorf("Members(nil, nil) = %v, want nil", got)
	}
	if got := ReturnType(nil, nil); got != nil {
		t.Errorf("ReturnType(nil, nil) = %v, want nil", got)
	}
	if got := TypeReferenceNames(nil); got != nil {
		t.Errorf("TypeReferenceNames(nil) = %v, want nil", got)
	}
	if got := WrittenName(nil); got != "" {
		t.Errorf("WrittenName(nil) = %q, want empty", got)
	}
}
