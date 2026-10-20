package globalipranges

import (
	tf "github.com/catonetworks/terraform-provider-cato/internal/provider/shared/tfmodel"
)

type ipRangePlanDetails struct {
	toCreate []tf.GlobalIPRange
	toUpdate []tf.GlobalIPRange
	toDelete []tf.GlobalIPRange
	toKeep   []tf.GlobalIPRange
}
