package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testVisitorID = "test_visitor_id"
var testEnvID = "test_env_id"

func TestValidateBase(t *testing.T) {
	b := BaseHit{}

	errs := b.validateBase()
	if len(errs) != 4 {
		t.Errorf("Empty hit should raise 4 errors. %d raised", len(errs))
	}

	b.VisitorID = testVisitorID

	errs = b.validateBase()
	if len(errs) != 3 {
		t.Errorf("Hit with only visitor ID set should raise 3 errors. %d raised", len(errs))
	}

	b.DataSource = "APP"
	errs = b.validateBase()
	if len(errs) != 2 {
		t.Errorf("Hit with empty type and missing environment ID should raise 2 errors. %d raised", len(errs))
	}

	b.EnvironmentID = testEnvID
	errs = b.validateBase()
	if len(errs) != 1 {
		t.Errorf("Hit with missing environment ID should raise 1 errors. %d raised", len(errs))
	}

	b.Type = "wrong_type"
	errs = b.validateBase()
	if len(errs) != 1 {
		t.Errorf("Hit with wrong type set should raise 1 errors. %d raised", len(errs))
	}

	b.Type = TRANSACTION
	errs = b.validateBase()
	if len(errs) != 0 {
		t.Errorf("Hit with mandatory fields set should not raise any errors. %d raised", len(errs))
	}
}

func TestValidatePage(t *testing.T) {
	b := PageHit{
		BaseHit: BaseHit{},
	}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	assert.Equal(t, 1, len(errs))

	b.DocumentLocation = "location"

	errs = b.Validate()
	assert.Equal(t, 0, len(errs))
}

func TestValidateScreen(t *testing.T) {
	b := ScreenHit{
		BaseHit: BaseHit{},
	}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	assert.Equal(t, 1, len(errs))

	b.PageTitle = "page title"

	errs = b.Validate()
	assert.Equal(t, 0, len(errs))
}

func TestValidateEvent(t *testing.T) {
	b := EventHit{
		BaseHit: BaseHit{},
	}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	if len(errs) != 1 {
		t.Errorf("Missing action event should raise 1 errors. %d raised", len(errs))
	}

	b.Action = "test"
	errs = b.Validate()
	if len(errs) != 0 {
		t.Errorf("Valid event hit should not raise any errors. %d raised", len(errs))
	}
}

func TestValidateTransaction(t *testing.T) {
	b := TransactionHit{
		BaseHit: BaseHit{},
	}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	if len(errs) != 2 {
		t.Errorf("Missing affiliation and id transaction should raise 2 errors. %d raised", len(errs))
	}

	b.TransactionID = "test_tid"
	errs = b.Validate()
	if len(errs) != 1 {
		t.Errorf("Missing affiliation should raise 1 errors. %d raised", len(errs))
	}

	b.Affiliation = "test_affiliation"
	errs = b.Validate()
	if len(errs) != 0 {
		t.Errorf("Correct transaction hit should not raise any errors. %d raised", len(errs))
	}
}

func TestValidateItem(t *testing.T) {
	b := ItemHit{
		BaseHit: BaseHit{},
	}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	if len(errs) != 3 {
		t.Errorf("Missing item name, code and id transaction should raise 3 errors. %d raised", len(errs))
	}

	b.TransactionID = "test_tid"
	errs = b.Validate()
	if len(errs) != 2 {
		t.Errorf("Missing affiliation should raise 2 errors. %d raised", len(errs))
	}

	b.Name = "test_item_name"
	errs = b.Validate()
	if len(errs) != 1 {
		t.Errorf("Correct transaction hit should raise 1 errors. %d raised", len(errs))
	}

	b.Code = "test_item_code"
	errs = b.Validate()
	if len(errs) != 0 {
		t.Errorf("Correct transaction hit should not raise any errors. %d raised", len(errs))
	}
}

func TestValidateActivation(t *testing.T) {
	b := ActivationHit{}
	b.SetBaseInfos(testEnvID, testVisitorID)

	errs := b.Validate()
	if len(errs) != 2 {
		t.Errorf("Missing campaign and variation id should raise 2 errors. %d raised", len(errs))
	}

	b.VariationGroupID = "test_vgid"
	errs = b.Validate()
	if len(errs) != 1 {
		t.Errorf("Missing variation id should raise 1 errors. %d raised", len(errs))
	}

	b.VariationID = "test_affiliation"
	errs = b.Validate()
	if len(errs) != 0 {
		t.Errorf("Correct activation hit should not raise any errors. %d raised", len(errs))
	}
}

func TestEvent(t *testing.T) {
	b := Event{}
	b.SetBaseInfos(testEnvID, "")
	errs := b.Validate()
	if len(errs) != 2 {
		t.Errorf("Wrong vID and and missing type should raise 2 errors. %d raised", len(errs))
	}

	b.SetBaseInfos(testEnvID, testVisitorID)

	errs = b.Validate()
	if len(errs) != 1 {
		t.Errorf("Missing type should raise 1 error. %d raised", len(errs))
	}

	b.Type = CONTEXT
	b.Data = Context{}
	errs = b.Validate()
	if len(errs) != 0 {
		t.Errorf("Correct activation hit should not raise any errors. %d raised", len(errs))
	}
}

func TestBatch(t *testing.T) {
	b := createBatchHit(BaseHit{})
	b.SetBaseInfos(testEnvID, testVisitorID)
	b.AddHit(&EventHit{Action: "event Action"})

	errs := b.Validate()
	if len(errs) != 0 {
		t.Errorf("Correct activation hit should not raise any errors. %d raised", len(errs))
	}

	assert.Equal(t, 1, len(b.Hits))

	time.Sleep(time.Second * 1)

	b.ComputeQueueTime()

	assert.GreaterOrEqual(t, b.QueueTime, int64(980))
	assert.LessOrEqual(t, b.QueueTime, int64(1020))
}
