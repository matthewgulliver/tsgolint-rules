package main

// The rules this repository owns.
//
// Registration is explicit and this list is the whole of it: nothing is picked
// up by scanning a directory, so a rule absent from here does not run however
// complete its package looks. `packages/oxlint/plugin.ts` is the same list for
// JavaScript half.
//
// Every rule of ours carries the tree it judges, declared beside it in its own
// package: `archlint` runs it over those files and no others.
//
// Upstream's own rules are off unless `.archtypesrc.json` asks for one by
// name. They name no tree, so any one of them judges every file the tsconfig
// includes — which is how `lint:arch-types` came to print 580 diagnostics that
// belonged to nobody. Naming one is a deliberate act, and `upstreamRules` is
// the list of the ones this repository has decided to have an opinion about.

import (
	"github.com/typescript-eslint/tsgolint/internal/archrule"
	"github.com/typescript-eslint/tsgolint/internal/rule"

	switch_exhaustiveness_check "github.com/typescript-eslint/tsgolint/internal/rules/switch_exhaustiveness_check"

	context_model_does_not_cross_the_boundary "github.com/typescript-eslint/tsgolint/internal/rules/context_model_does_not_cross_the_boundary"
	domain_function_returns_an_answer "github.com/typescript-eslint/tsgolint/internal/rules/domain_function_returns_an_answer"
	domain_probe_returns_void "github.com/typescript-eslint/tsgolint/internal/rules/domain_probe_returns_void"
	domain_signature_stays_in_the_domain "github.com/typescript-eslint/tsgolint/internal/rules/domain_signature_stays_in_the_domain"
	domain_state_is_deeply_readonly "github.com/typescript-eslint/tsgolint/internal/rules/domain_state_is_deeply_readonly"
	domain_type_is_declared_once "github.com/typescript-eslint/tsgolint/internal/rules/domain_type_is_declared_once"
	driving_port_command_is_modelled "github.com/typescript-eslint/tsgolint/internal/rules/driving_port_command_is_modelled"
	no_constructed_collaborators "github.com/typescript-eslint/tsgolint/internal/rules/no_constructed_collaborators"
	no_double_library_in_domain_test "github.com/typescript-eslint/tsgolint/internal/rules/no_double_library_in_domain_test"
	no_double_library_in_use_case_test "github.com/typescript-eslint/tsgolint/internal/rules/no_double_library_in_use_case_test"
	no_outside_declaration_in_the_hexagon "github.com/typescript-eslint/tsgolint/internal/rules/no_outside_declaration_in_the_hexagon"
	no_page_request_in_journey "github.com/typescript-eslint/tsgolint/internal/rules/no_page_request_in_journey"
	no_provider_type_in_signature "github.com/typescript-eslint/tsgolint/internal/rules/no_provider_type_in_signature"
	port_behaviour_is_an_interface "github.com/typescript-eslint/tsgolint/internal/rules/port_behaviour_is_an_interface"
	published_contract_publishes_no_mutable_value "github.com/typescript-eslint/tsgolint/internal/rules/published_contract_publishes_no_mutable_value"
	read_port_returns_an_answer "github.com/typescript-eslint/tsgolint/internal/rules/read_port_returns_an_answer"
	stored_state_switch_has_a_throwing_default "github.com/typescript-eslint/tsgolint/internal/rules/stored_state_switch_has_a_throwing_default"
	use_case_result_is_discriminated "github.com/typescript-eslint/tsgolint/internal/rules/use_case_result_is_discriminated"
	use_case_throws_a_domain_error "github.com/typescript-eslint/tsgolint/internal/rules/use_case_throws_a_domain_error"
)

var archRules = []archrule.Rule{
	context_model_does_not_cross_the_boundary.ContextModelDoesNotCrossTheBoundaryRule,
	domain_function_returns_an_answer.DomainFunctionReturnsAnAnswerRule,
	domain_probe_returns_void.DomainProbeReturnsVoidRule,
	domain_signature_stays_in_the_domain.DomainSignatureStaysInTheDomainRule,
	domain_state_is_deeply_readonly.DomainStateIsDeeplyReadonlyRule,
	domain_type_is_declared_once.DomainTypeIsDeclaredOnceRule,
	driving_port_command_is_modelled.DrivingPortCommandIsModelledRule,
	no_constructed_collaborators.NoConstructedCollaboratorsRule,
	no_double_library_in_domain_test.NoDoubleLibraryInDomainTestRule,
	no_double_library_in_use_case_test.NoDoubleLibraryInUseCaseTestRule,
	no_outside_declaration_in_the_hexagon.NoOutsideDeclarationInTheHexagonRule,
	no_page_request_in_journey.NoPageRequestInJourneyRule,
	no_provider_type_in_signature.NoProviderTypeInSignatureRule,
	port_behaviour_is_an_interface.PortBehaviourIsAnInterfaceRule,
	published_contract_publishes_no_mutable_value.PublishedContractPublishesNoMutableValueRule,
	read_port_returns_an_answer.ReadPortReturnsAnAnswerRule,
	stored_state_switch_has_a_throwing_default.StoredStateSwitchHasAThrowingDefaultRule,
	use_case_result_is_discriminated.UseCaseResultIsDiscriminatedRule,
	use_case_throws_a_domain_error.UseCaseThrowsADomainErrorRule,
}

// Upstream rules this repository carries. Off by default; a configuration file
// turns one on. `docs/rules/README.md` explains why the exhaustiveness half of
// "returns a discriminated union and handles it exhaustively" is upstream's job
// and not ours.
var upstreamRules = []rule.Rule{
	switch_exhaustiveness_check.SwitchExhaustivenessCheckRule,
}
