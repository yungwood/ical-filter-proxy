import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import CodeBlock from '@theme/CodeBlock';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';

import styles from './index.module.css';

function HomepageHeader() {
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <img
          src={require('../../static/img/logo.png').default}
          alt="iCal Filter Proxy logo"
          className={styles.heroLogo}
        />
        <Heading as="h1" className="hero__title">
          Clean up noisy calendar feeds
        </Heading>
        <p className="hero__subtitle">
          Publish filtered iCalendar feeds from upstream calendars using
          config-driven match and transform rules.
        </p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/getting-started">
            Quick Start
          </Link>
          <Link
            className="button button--outline button--secondary button--lg"
            to="/configuration">
            Configuration
          </Link>
        </div>
      </div>
    </header>
  );
}

type FeatureSectionProps = {
  title: string;
  body: ReactNode;
  links: Array<{to: string; label: string}>;
  code: string;
  language: string;
  codeTitle: string;
  shaded?: boolean;
  reverse?: boolean;
};

function FeatureSection({
  title,
  body,
  links,
  code,
  language,
  codeTitle,
  shaded,
  reverse,
}: FeatureSectionProps): ReactNode {
  const textColumn = (
    <div className={clsx('col col--5', styles.textColumn)}>
      <div className={styles.textContent}>
        <Heading as="h2">{title}</Heading>
        <p>{body}</p>
        <div className={styles.sectionLinks}>
          {links.map((link) => (
            <Link key={link.to} to={link.to} className={styles.sectionLink}>
              <span>{link.label}</span>
              <span aria-hidden="true">→</span>
            </Link>
          ))}
        </div>
      </div>
    </div>
  );
  const codeColumn = (
    <div className={clsx('col col--7', styles.codeColumn)}>
      <div className={styles.codeContent}>
        <CodeBlock title={codeTitle} language={language} className={styles.codeBlock}>
          {code}
        </CodeBlock>
      </div>
    </div>
  );

  return (
    <section
      className={clsx(
        styles.featureSection,
        shaded && styles.shadedSection,
        reverse && styles.reverseSection,
      )}>
      <div className="container">
        <div className="row">
          {reverse ? codeColumn : textColumn}
          {reverse ? textColumn : codeColumn}
        </div>
      </div>
    </section>
  );
}

function HomepageSections(): ReactNode {
  return (
    <>
      <FeatureSection
        title="Proxy upstream calendars"
        body="Publish stable feed URLs backed by upstream iCalendar sources, with per-calendar auth and output naming."
        links={[
          {to: '/configuration/calendars', label: 'Calendar configuration'},
          {to: '/configuration/authentication', label: 'Authentication modes'},
        ]}
        language="yaml"
        codeTitle="config.yaml"
        code={`calendars:
  - name: work
    publish_name: "Work Calendar"
    token: "changeme"
    feed_url: "https://example.com/calendar.ics"
    user_agent: "ical-filter-proxy"`}
      />
      <FeatureSection
        title="Filter and transform events"
        body="Define ordered rules that match event fields, remove unwanted events, and rewrite noisy summaries before publishing the feed."
        links={[
          {to: '/filtering', label: 'Filtering overview'},
          {to: '/recipes/remove-cancelled-events', label: 'Example: Remove cancelled events'},
          {to: '/recipes/keep-only-matching-events', label: 'Example: Keep only matching events'},
        ]}
        language="yaml"
        codeTitle="config.yaml"
        shaded
        reverse
        code={`calendars:
  - name: work
    public: true
    feed_url: "https://example.com/calendar.ics"
    filters:
      - description: "Remove cancelled events"
        remove: true
        match:
          summary:
            prefix: "Canceled: "
      - description: "Clean summary text"
        transform:
          summary:
            trim_prefix: "Roster - "
            replace_text:
              old: "Primary On Call"
              new: "On-Call"
              all: true`}
      />
      <FeatureSection
        title="Deploy and operate"
        body="Run with first-class Docker and Helm support, including secret-file config for tokens and upstream URLs, management endpoints, metrics, and structured logs."
        links={[
          {to: '/installation/docker', label: 'Docker and Compose'},
          {to: '/installation/kubernetes-helm', label: 'Kubernetes / Helm'},
          {to: '/configuration/secret-files', label: 'Secret files'},
          {to: '/operations/metrics', label: 'Metrics'},
        ]}
        language="yaml"
        codeTitle="values.yaml"
        code={`metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    labels:
      release: prometheus
    interval: 30s`}
      />
    </>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="iCal proxy with support for user-defined filtering rules">
      <HomepageHeader />
      <main>
        <HomepageSections />
      </main>
    </Layout>
  );
}
