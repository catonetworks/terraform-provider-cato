// Assign static license to site
resource "cato_license" "site-static" {
  site_id = "123456"
  license_id = "abcde-1234-abcd-1234-abcde1234"
}

// Assign Pooled license using data source to retrieve license id
data "cato_licensingInfo" "pb_license" {
  is_active = true
  sku = "CATO_PB"
}

resource "cato_license" "site-pooled" {
  site_id = "123456"
  license_id = data.cato_licensingInfo.pb_license.licenses[0].id
  bw = 40
}

