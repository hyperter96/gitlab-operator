# https://guides.rubyonrails.org/action_mailer_basics.html#action-mailer-configuration
if Rails.env.production?
    Rails.application.config.action_mailer.delivery_method = :smtp
  
    ActionMailer::Base.delivery_method = :smtp
    ActionMailer::Base.smtp_settings = {
      {{ if .Host }}address: "{{ .Host }}",{{ end }}
      {{ if .Port }}port: {{ .Port }},{{ end }}
      {{ if .Username }}user_name: "{{ .Username }}",{{ end }}
      {{ if .Password }}password: "{{ .Password }}",{{ end }}
      {{ if .Domain }}domain: "{{ .Domain }}",{{ end }}
      {{ if .Authentication }}authentication: :{{ .Authentication }},{{ end }}
      {{ if .EnableStartTLS }}enable_starttls_auto: {{ .EnableStartTLS }},{{ end }}
      {{ if .OpenSSLVerifyMode }}openssl_verify_mode: '{{ .OpenSSLVerifyMode }}',{{ end }}
      {{ if .EnableSSL }}tls: {{ .EnableSSL }},{{ end }}
    }
end
