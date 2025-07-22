{
  config,
  lib,
  pkgs,
  ...
}:
with lib; let
  cfg = config.services.ical-filter-proxy;
in {
  options.services.ical-filter-proxy = {
    enable = mkEnableOption "iCal Filter Proxy service";

    package = mkOption {
      type = types.package;
      default = pkgs.ical-filter-proxy;
      defaultText = literalExpression "pkgs.ical-filter-proxy";
      description = "The ical-filter-proxy package to use.";
    };

    user = mkOption {
      type = types.str;
      default = "ical-filter-proxy";
      description = "User account under which ical-filter-proxy runs.";
    };

    group = mkOption {
      type = types.str;
      default = "ical-filter-proxy";
      description = "Group under which ical-filter-proxy runs.";
    };

    port = mkOption {
      type = types.port;
      default = 8080;
      description = "Port on which the service listens.";
    };

    configFile = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = "Path to the configuration file for ical-filter-proxy. Either this or config must be specified.";
    };

    config = mkOption {
      type = types.nullOr (types.submodule {
        options = {
          calendars = mkOption {
            type = types.listOf (types.submodule {
              options = {
                name = mkOption {
                  type = types.str;
                  description = "Name of the calendar (used in URL path).";
                };
                publish_name = mkOption {
                  type = types.nullOr types.str;
                  default = null;
                  description = "Published name of the calendar. Uses upstream value if not specified.";
                };
                token = mkOption {
                  type = types.nullOr types.str;
                  default = null;
                  description = "Authentication token for accessing the calendar feed.";
                };
                token_file = mkOption {
                  type = types.nullOr types.path;
                  default = null;
                  description = "Path to file containing the authentication token.";
                };
                public = mkOption {
                  type = types.bool;
                  default = false;
                  description = "Whether the calendar is public (no authentication required).";
                };
                feed_url = mkOption {
                  type = types.nullOr types.str;
                  default = null;
                  description = "URL of the upstream iCal feed.";
                };
                feed_url_file = mkOption {
                  type = types.nullOr types.path;
                  default = null;
                  description = "Path to file containing the feed URL.";
                };
                filters = mkOption {
                  type = types.listOf (types.submodule {
                    options = {
                      description = mkOption {
                        type = types.str;
                        description = "Human-readable description of the filter.";
                      };
                      remove = mkOption {
                        type = types.bool;
                        default = false;
                        description = "Whether to remove events matching this filter.";
                      };
                      stop = mkOption {
                        type = types.bool;
                        default = false;
                        description = "Whether to stop processing further filters after this one matches.";
                      };
                      match = mkOption {
                        type = types.nullOr (types.submodule {
                          options = {
                            summary = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  empty = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Match if summary is empty.";
                                  };
                                  contains = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if summary contains this string.";
                                  };
                                  prefix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if summary starts with this string.";
                                  };
                                  suffix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if summary ends with this string.";
                                  };
                                  regex = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if summary matches this regex.";
                                  };
                                };
                              });
                              default = null;
                              description = "Match conditions for event summary.";
                            };
                            location = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  empty = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Match if location is empty.";
                                  };
                                  contains = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if location contains this string.";
                                  };
                                  prefix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if location starts with this string.";
                                  };
                                  suffix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if location ends with this string.";
                                  };
                                  regex = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if location matches this regex.";
                                  };
                                };
                              });
                              default = null;
                              description = "Match conditions for event location.";
                            };
                            description = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  empty = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Match if description is empty.";
                                  };
                                  contains = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if description contains this string.";
                                  };
                                  prefix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if description starts with this string.";
                                  };
                                  suffix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if description ends with this string.";
                                  };
                                  regex = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if description matches this regex.";
                                  };
                                };
                              });
                              default = null;
                              description = "Match conditions for event description.";
                            };
                            url = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  empty = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Match if URL is empty.";
                                  };
                                  contains = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if URL contains this string.";
                                  };
                                  prefix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if URL starts with this string.";
                                  };
                                  suffix = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if URL ends with this string.";
                                  };
                                  regex = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Match if URL matches this regex.";
                                  };
                                };
                              });
                              default = null;
                              description = "Match conditions for event URL.";
                            };
                          };
                        });
                        default = null;
                        description = "Conditions for matching events.";
                      };
                      transform = mkOption {
                        type = types.nullOr (types.submodule {
                          options = {
                            summary = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  replace = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Replace summary with this string.";
                                  };
                                  remove = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Remove summary (set to empty string).";
                                  };
                                };
                              });
                              default = null;
                              description = "Transformations for event summary.";
                            };
                            location = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  replace = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Replace location with this string.";
                                  };
                                  remove = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Remove location (set to empty string).";
                                  };
                                };
                              });
                              default = null;
                              description = "Transformations for event location.";
                            };
                            description = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  replace = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Replace description with this string.";
                                  };
                                  remove = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Remove description (set to empty string).";
                                  };
                                };
                              });
                              default = null;
                              description = "Transformations for event description.";
                            };
                            url = mkOption {
                              type = types.nullOr (types.submodule {
                                options = {
                                  replace = mkOption {
                                    type = types.nullOr types.str;
                                    default = null;
                                    description = "Replace URL with this string.";
                                  };
                                  remove = mkOption {
                                    type = types.nullOr types.bool;
                                    default = null;
                                    description = "Remove URL (set to empty string).";
                                  };
                                };
                              });
                              default = null;
                              description = "Transformations for event URL.";
                            };
                          };
                        });
                        default = null;
                        description = "Transformations to apply to matching events.";
                      };
                    };
                  });
                  default = [];
                  description = "List of filters to apply to this calendar.";
                };
              };
            });
            description = "List of calendars to proxy and filter.";
          };
        };
      });
      default = null;
      example = {
        calendars = [
          {
            name = "example";
            token = "secure-token";
            feed_url = "https://calendar.example.com/feed.ics";
            filters = [
              {
                description = "Remove canceled events";
                remove = true;
                match.summary.prefix = "Canceled: ";
              }
            ];
          }
        ];
      };
      description = "Configuration as a Nix attribute set. Either this or configFile must be specified.";
    };

    debug = mkOption {
      type = types.bool;
      default = false;
      description = "Enable debug logging.";
    };

    jsonLogging = mkOption {
      type = types.bool;
      default = false;
      description = "Output logging in JSON format.";
    };

    extraArgs = mkOption {
      type = types.listOf types.str;
      default = [];
      description = "Additional command line arguments to pass to ical-filter-proxy.";
    };

    openFirewall = mkOption {
      type = types.bool;
      default = false;
      description = "Whether to open the firewall for the specified port.";
    };
  };

  config = mkIf cfg.enable {
    assertions = [
      {
        assertion = (cfg.configFile != null) != (cfg.config != null);
        message = "Either configFile or config must be specified, but not both.";
      }
    ];

    users.users.${cfg.user} = {
      description = "iCal Filter Proxy service user";
      isSystemUser = true;
      group = cfg.group;
    };

    users.groups.${cfg.group} = {};

    systemd.services.ical-filter-proxy = let
      configFile =
        if cfg.configFile != null
        then cfg.configFile
        else pkgs.writeText "ical-filter-proxy-config.yaml" (pkgs.lib.generators.toYAML {} cfg.config);
    in {
      description = "iCal Filter Proxy";
      after = ["network.target"];
      wantedBy = ["multi-user.target"];

      serviceConfig = {
        Type = "exec";
        User = cfg.user;
        Group = cfg.group;
        Restart = "on-failure";
        RestartSec = "5s";

        # Security settings
        NoNewPrivileges = true;
        PrivateTmp = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectControlGroups = true;
        RestrictAddressFamilies = ["AF_UNIX" "AF_INET" "AF_INET6"];
        RestrictNamespaces = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        RemoveIPC = true;
        CapabilityBoundingSet = "";

        # Allow reading the config file
        ReadOnlyPaths = [configFile];

        ExecStart = let
          args =
            [
              "${cfg.package}/bin/ical-filter-proxy"
              "-config"
              "${configFile}"
              "-port"
              "${toString cfg.port}"
            ]
            ++ optional cfg.debug "-debug"
            ++ optional cfg.jsonLogging "-json"
            ++ cfg.extraArgs;
        in "${lib.escapeShellArgs args}";
      };
    };

    networking.firewall = mkIf cfg.openFirewall {
      allowedTCPPorts = [cfg.port];
    };
  };
}
