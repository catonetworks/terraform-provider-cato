# References existing destinations. This does not attach an event producer or send notifications.
resource "cato_subscription_group" "example" {
  name         = "Operations notifications"
  mailing_list = [{ id = "123" }]
  # integration = [{ name = "Existing webhook" }]
}

# Creation requires at least one mailing list or integration.
# Removing a membership collection clears its currently observed members.
# terraform import cato_subscription_group.example 12345
