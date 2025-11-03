# Create section for WAN Network Rules rules 
resource "cato_wnw_section" "section1" {
    at = {
        position = "LAST_IN_POLICY"
    }
    section = {
        name = "Custom WAN Network Rules"
    }    
}
