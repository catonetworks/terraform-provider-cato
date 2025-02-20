// internet firewall allowing all & logs
resource "cato_if_rule" "allow_all_and_log" {
  at = {
    position = "LAST_IN_POLICY"
  }
  rule = {
    name    = "Allow all & logs"
    enabled = true
    action  = "ALLOW"
    tracking = {
      event = {
        enabled = true
      }
    }
  }
}

// block all remote users except "Marketing" using category domain "test.com"
resource "cato_if_rule" "block_test_com_for_remote_users" {
  at = {
    position = "FIRST_IN_POLICY"
  }
  rule = {
    name              = "Block Test.com for Remote Users"
    enabled           = true
    action            = "BLOCK"
    connection_origin = "REMOTE"
    destination = {
      domain = [
        "test.com"
      ]
    }
    source = {}
    tracking = {
      event = {
        enabled = true
      }
    }
    exceptions = [
      {
        name = "Exclude Marketing Teams"
        source = {
          users_group = [
            {
              name = "Marketing-Teams"
            }
          ]
        }
      }
    ]
  }
}

