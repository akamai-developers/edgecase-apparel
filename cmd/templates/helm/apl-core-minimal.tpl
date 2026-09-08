{{- $apl := .main }}
{{- $objKeys := .opts.objKeys }}
{{- $objBuckets := .opts.objBuckets }}
cluster:
  name: {{ $apl.PlatformName }}
  domainSuffix: {{ $apl.Domain }}
  provider: linode
otomi:
  adminPassword: {{ randInitPass }}
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
      region: {{ $apl.Region }}
      accessKeyId: {{ $objKeys.accessKey }}
      secretAccessKey: {{ $objKeys.secretKey }}
      buckets:
        {{- range $key, $value := $objBuckets }}
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
