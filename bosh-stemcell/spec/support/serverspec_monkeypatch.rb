require 'serverspec'

# `example` method monkey path
unless Specinfra::VERSION == '2.67.1'
  raise "Unexpected Specinfra version #{Specinfra::VERSION}"
end

# Exec monkey path
require 'monkeypatch/serverspec/backend/exec'
# include Serverspec::Helper::Exec
# include Serverspec::Helper::DetectOS
set :backend, :exec
