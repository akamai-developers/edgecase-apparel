{{- $apl := .main }}
{{- $objKeys := .opts.objKeys }}
{{- $objBuckets := .opts.objBuckets }}
{{- $objBuckets := .opts.objBuckets }}
{{- $apps := .opts.apps }}
{{- $keycloak := .opts.keycloak }}
{{- $teams := .opts.teams }}
{{- $git := .opts.git }}
catalogs:
  product:
    name: core
    repositoryUrl: https://github.com/akamai-developers-bot/apl-charts
    branch: main
    enabled: true
  systems:
    name: sysdev
    repositoryUrl: https://github.com/ruckus-voxi/apl-charts
    branch: main
    enabled: true
cluster:
  name: {{ $apl.PlatformName }}
  domainSuffix: {{ $apl.Domain }}
  provider: linode
otomi:
  git:
    repoUrl: https://github.com/ruckus-voxi/edgecase-apparel-platform
    username: ruckus-voxi
    password: {{ $git.ghtoken }}
    email: rvoxi@akamai.com
    branch: v0.1.0
  adminPassword: {{ $keycloak.adminPass }}
  hasExternalDNS: true
dns:
  provider:
    linode:
      apiToken: {{ $apl.Token }}
  domainFilters: 
    - {{ $apl.Domain }}
apps:
  alertmanager:
    enabled: true
  cert-manager:
    issuer: letsencrypt
    stage: production
    email: admin@{{ $apl.Domain }}
  gitea:
    enabled: true
  grafana:
    enabled: true
  harbor:
    enabled: true
  keycloak:
    idp:
      clientSecret: {{ $keycloak.clientSecret }}
  kserve:
    enabled: true
  kyverno:
    enabled: true
  kubeflow-pipelines:
    enabled: true
  loki:
    enabled: true
    adminPassword: {{ $apps.loki }}
  otel:
    enabled: true
  prometheus:
    enabled: true
  promtail:
    enabled: true
  rabbitmq:
    enabled: true
  trivy:
    enabled: true
obj:
  showWizard: false
  provider:
    type: linode
    linode:
      region: {{ $apl.Region }}
      accessKeyId: {{ $objKeys.accessKey }}
      secretAccessKey: {{ $objKeys.secretKey }}
      buckets:
        {{- range $key, $value := $objBuckets }}
        {{ $key }}: {{ $value }}
        {{- end }}
platformBackups:
  database:
    harbor:
      enabled: true
      pathSuffix: harbor
      retentionPolicy: 7d
      schedule: 0 0 * * *
    gitea:
      enabled: true
      pathSuffix: gitea
      retentionPolicy: 7d
      schedule: 0 1 * * *
    keycloak:
      enabled: true
      pathSuffix: keycloak
      retentionPolicy: 7d
      schedule: 0 2 * * *
  gitea:
    enabled: true
    retentionPolicy: 7d
    schedule: 0 3 * * *
teamConfig:
  core:
    settings:
      password: {{ $teams.core }}
      id: core
      selfService:
        teamMembers:
          createServices: true
          editSecurityPolicies: false
          useCloudShell: true
          downloadKubeconfig: true
          downloadDockerLogin: false
      managedMonitoring:
        grafana: true
        alertmanager: true
      networkPolicy:
        egressPublic: true
        ingressPrivate: true
  sysdev:
    settings:
      password: {{ $teams.sysdev }}
      id: sysdev
      selfService:
        teamMembers:
          createServices: true
          editSecurityPolicies: true
          useCloudShell: true
          downloadKubeconfig: true
          downloadDockerLogin: true
      managedMonitoring:
        grafana: true
        alertmanager: true
      networkPolicy:
        egressPublic: true
        ingressPrivate: true
users:
  - email: ruckus@{{ $apl.Domain }}
    firstName: Ruckus
    lastName: Voxi
    isPlatformAdmin: true
    isTeamAdmin: false
    teams: []
    initialPassword: {{ randInitPass }}
  - email: duan@{{ $apl.Domain }}
    firstName: Du'An
    lastName: Lightfoot
    isPlatformAdmin: false
    isTeamAdmin: true
    teams:
      - core
    initialPassword: {{ randInitPass }}
  - email: sheilah@{{ $apl.Domain }}
    firstName: Sheilah
    lastName: Kirui
    isPlatformAdmin: false
    isTeamAdmin: true
    teams:
      - core
    initialPassword: {{ randInitPass }}
  - email: thorsten@{{ $apl.Domain }}
    firstName: Thorsten
    lastName: Hans
    isPlatformAdmin: false
    isTeamAdmin: true
    teams:
      - core
      - sysdev
    initialPassword: {{ randInitPass }}