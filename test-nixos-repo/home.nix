# Home Manager Configuration
{
  config,
  pkgs,
  ...
}: {
  # Home Manager needs a bit of information about you and the
  # paths it should manage.
  home.username = "user";
  home.homeDirectory = "/home/user";

  # This value determines the Home Manager release that your
  # configuration is compatible with.
  home.stateVersion = "23.11";

  # Let Home Manager install and manage itself.
  programs.home-manager.enable = true;

  # Enable Git
  programs.git = {
    enable = true;
    userName = "User Name";
    userEmail = "user@example.com";
  };

  # Enable Zsh
  programs.zsh = {
    enable = true;
    enableCompletion = true;
    syntaxHighlighting.enable = true;
  };

  # Install user packages
  home.packages = with pkgs; [
    firefox
    vscode
    neovim
    tmux
    fzf
  ];

  # Configure Neovim
  programs.neovim = {
    enable = true;
    viAlias = true;
    vimAlias = true;
  };
}
