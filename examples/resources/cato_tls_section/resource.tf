# Create section for custom TLS rules 
resource "cato_tls_section" "section1" {
    at = {
        position = "LAST_IN_POLICY"
    }
    section = {
        name = "Custom TLS Rules"
    }    
}

# Create first section for TLS phases
resource "cato_tls_section" "section2" {
    at      = {
        position = "AFTER_SECTION"
        ref = cato_tls_section.section1.section.id
    }
    section = {
        name = "TLS Phases"
    }    
}
