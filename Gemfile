source 'https://rubygems.org'

group :development, :test do
  gem 'rake', '~>10.0'
  gem 'rspec'
  gem 'rspec-its'
  gem 'rspec-instafail'
  gem 'bosh-stemcell', path: 'bosh-stemcell'
  gem 'bosh-core'
  gem 'bosh-dev', path: 'bosh-dev'
  gem 'serverspec', '~> 2.36.0'
  # Explicitly do not require serverspec dependency
  # so that it could be monkey patched in a deterministic way
  # in `bosh-stemcell/spec/support/serverspec_monkeypatch.rb`
  gem 'logging'
end
