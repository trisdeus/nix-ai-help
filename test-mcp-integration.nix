# Test file for MCP integration
# Try asking an AI assistant (Claude Dev or GitHub Copilot) about these options

{
  # Nginx web server configuration
  services.nginx = {
    enable = true;  # What does this option do?
    
    # What are recommended virtual host settings?
    virtualHosts."example.com" = {
      root = "/var/www/example";
      locations."/" = {
        index = "index.html";
      };
    };
  };
  
  # What does this Home Manager option do?
  programs.git = {
    enable = true;
    userName = "Example User";
    userEmail = "user@example.com";
  };
}
