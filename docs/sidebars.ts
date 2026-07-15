import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'getting-started/index',
    {
      type: 'category',
      label: 'Installation',
      items: [
        'installation/docker',
        'installation/kubernetes-helm',
        'installation/nix',
        'installation/source',
      ],
    },
    {
      type: 'category',
      label: 'Configuration',
      link: {type: 'doc', id: 'configuration/index'},
      items: [
        'configuration/calendars',
        'configuration/authentication',
        'configuration/environment-variables',
        'configuration/secret-files',
        'configuration/environment-substitution',
        'configuration/complete-example',
      ],
    },
    {
      type: 'category',
      label: 'Filtering',
      link: {type: 'doc', id: 'filtering/index'},
      items: [
        'filtering/match-rules',
        'filtering/transforms',
        'filtering/rule-ordering',
      ],
    },
    {
      type: 'category',
      label: 'Recipes',
      items: [
        'recipes/remove-cancelled-events',
        'recipes/remove-optional-events',
        'recipes/remove-public-holidays',
        'recipes/remove-empty-descriptions',
        'recipes/remove-outlook-following-events',
        'recipes/remove-outlook-private-appointments',
        'recipes/keep-only-matching-events',
        'recipes/strip-summary-text',
        'recipes/hide-event-descriptions',
        'recipes/set-custom-user-agent',
      ],
    },
    {
      type: 'category',
      label: 'Operations',
      items: [
        'operations/endpoints',
        'operations/metrics',
        'operations/logging',
        'operations/health-readiness',
        'operations/reverse-proxies',
        'operations/http-behavior',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/cli-flags',
        'reference/config-reference',
        'reference/metrics-reference',
      ],
    },
  ],
};

export default sidebars;
