cluster:
  name: {{ .PlatformName }}
  provider: linode
  domainSuffix: {{ .Domain }}
otomi:
  adminPassword: {{ randInitPass }}
  hasExternalDNS: true
dns:
  domainFilters: 
    - {{ .Domain }}
  provider:
    linode:
      apiToken: {{ .Token }}
apps:
  alertmanager:
    enabled: true
  cert-manager:
    issuer: letsencrypt
    stage: production
    email: admin@{{ .Domain }}
  grafana:
    enabled: true
  harbor:
    enabled: true
  loki:
    enabled: true
    adminPassword: {{ randInitPass }}
  prometheus:
    enabled: true
  tempo:
    enabled: true
obj:
  provider:
    type: linode
    linode:
      region: {{ .Region }}
      accessKeyId: {{ .accessKey }}
      secretAccessKey: {{ .secretKey }}
      buckets:
        {{- range $key, $value := .objBuckets }}
        {{ $key }}: {{ $value }}
        {{- end }}
teamConfig:
  develop:
    settings:
      password: {{ randInitPass }}
      id: develop
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
