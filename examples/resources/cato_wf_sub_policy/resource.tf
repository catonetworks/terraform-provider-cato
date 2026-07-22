// WAN Firewall sub-policy (a nested policy scoped by a SUB_POLICY_SCOPE rule).
//
// NOTE: The Cato API has no updateSubPolicy mutation, so changing `name` or
// `description` forces replacement, which also removes any rules contained in
// the sub-policy. The scope must set at least one non-ANY match field.

// minimal sub-policy
resource "cato_wf_sub_policy" "minimal" {
  name        = "Branch Sites Sub-Policy"
  description = "Rules that only apply to branch sites"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled     = true
    source      = {}
    destination = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// full sub-policy with a scoped source and a nested rule owned by it
resource "cato_wf_sub_policy" "full" {
  name        = "Datacenter Sub-Policy"
  description = "Rules scoped to datacenter hosts"

  at = {
    position = "LAST_IN_POLICY"
  }

  scope = {
    enabled = true
    source = {
      ip = ["10.0.0.0/8"]
    }
    destination = {}
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// a rule placed inside the "full" sub-policy via sub_policy_id
resource "cato_wf_rule" "in_sub_policy" {
  sub_policy_id = cato_wf_sub_policy.full.id

  // Rules owned by a sub-policy are always positioned before the sub-policy
  // cleanup rule; the `at` position below is required by the schema but the
  // provider anchors the rule inside the sub-policy automatically.
  at = {
    position = "LAST_IN_POLICY"
  }

  rule = {
    name    = "Allow datacenter to internal"
    enabled = true
    action  = "ALLOW"
    source  = {}
    destination = {
      ip = ["10.1.0.0/16"]
    }
    tracking = {
      event = {
        enabled = true
      }
    }
  }

  depends_on = [cato_wf_sub_policy.full]
}
