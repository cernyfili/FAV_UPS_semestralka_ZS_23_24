package internal

import (
	"context"
	"fmt"
	"gameserver/pkg/stateless"
	"reflect"
)

const (
	triggerCallDialed             = "CallDialed"
	triggerCallConnected          = "CallConnected"
	triggerLeftMessage            = "LeftMessage"
	triggerPlacedOnHold           = "PlacedOnHold"
	triggerTakenOffHold           = "TakenOffHold"
	triggerPhoneHurledAgainstWall = "PhoneHurledAgainstWall"
	triggerMuteMicrophone         = "MuteMicrophone"
	triggerUnmuteMicrophone       = "UnmuteMicrophone"
	triggerSetVolume              = "SetVolume"
)

const (
	stateOffHook        = "OffHook"
	stateRinging        = "Ringing"
	stateConnected      = "Connected"
	stateOnHold         = "OnHold"
	statePhoneDestroyed = "PhoneDestroyed"
)

func Example() {
	phoneCall := stateless.NewStateMachine(stateOffHook)

	phoneCall.SetTriggerParameters(triggerSetVolume, reflect.TypeOf(0))
	phoneCall.SetTriggerParameters(triggerCallDialed, reflect.TypeOf(""))

	phoneCall.Configure(stateOffHook).
		Permit(triggerCallDialed, stateRinging)

	phoneCall.Configure(stateRinging).
		OnEntryFrom(triggerCallDialed, func(_ context.Context, args ...any) error {
			onDialed(args[0].(string))
			return nil
		}).
		Permit(triggerCallConnected, stateConnected)

	phoneCall.Configure(stateConnected).
		OnEntry(startCallTimer).
		OnExit(func(_ context.Context, _ ...any) error {
			stopCallTimer()
			return nil
		}).
		InternalTransition(triggerMuteMicrophone, func(_ context.Context, _ ...any) error {
			onMute()
			return nil
		}).
		InternalTransition(triggerUnmuteMicrophone, func(_ context.Context, _ ...any) error {
			onUnmute()
			return nil
		}).
		InternalTransition(triggerSetVolume, func(_ context.Context, args ...any) error {
			onSetVolume(args[0].(int))
			return nil
		}).
		Permit(triggerLeftMessage, stateOffHook).
		Permit(triggerPlacedOnHold, stateOnHold)

	phoneCall.Configure(stateOnHold).
		SubstateOf(stateConnected).
		Permit(triggerTakenOffHold, stateConnected).
		Permit(triggerPhoneHurledAgainstWall, statePhoneDestroyed)

	phoneCall.ToGraph()

	/*phoneCall.OnUnhandledTrigger(func(_ context.Context, state string, _ string, _ []string) {})*/

	err := phoneCall.Fire(triggerCallDialed, "qmuntal")
	if err != nil {
		return
	}
	phoneCall.Fire(triggerCallConnected)
	phoneCall.Fire(triggerSetVolume, 2)
	phoneCall.Fire(triggerPlacedOnHold)
	phoneCall.Fire(triggerMuteMicrophone)
	phoneCall.Fire(triggerUnmuteMicrophone)
	phoneCall.Fire(triggerTakenOffHold)
	phoneCall.Fire(triggerSetVolume, 11)
	phoneCall.Fire(triggerPlacedOnHold)
	phoneCall.Fire(triggerPhoneHurledAgainstWall)
	fmt.Printf("State is %v\n", phoneCall.MustState())

	/*	isValid, err := phoneCall.CanFire(triggerTakenOffHold)
		fmt.Printf("CanFire: %v, err: %v\n", isValid, err)*/

	err = phoneCall.Fire(triggerTakenOffHold)
	fmt.Printf("Fire: err: %v\n", err)

	// Output:
	// [Phone Call] placed for : [qmuntal]
	// [Timer:] Call started at 11:00am
	// Volume set to 2!
	// Microphone muted!
	// Microphone unmuted!
	// Volume set to 11!
	// [Timer:] Call ended at 11:30am
	// State is PhoneDestroyed

}

func handleError(context context.Context, state string, trigger string, args []string) {
	fmt.Printf("Wrong trigger on state %s: %s\n", state, trigger)

	print(context)
	print(args)
}

func onSetVolume(volume int) {
	fmt.Printf("Volume set to %d!\n", volume)
}

func onUnmute() {
	fmt.Println("Microphone unmuted!")
}

func onMute() {
	fmt.Println("Microphone muted!")
}

func onDialed(callee string) {
	fmt.Printf("[Phone Call] placed for : [%s]\n", callee)
}

func startCallTimer(_ context.Context, _ ...any) error {
	fmt.Println("[Timer:] Call started at 11:00am")
	return nil
}

func stopCallTimer() {
	fmt.Println("[Timer:] Call ended at 11:30am")
}
