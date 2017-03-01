require 'rspec'
require 'rspec/its'
# require 'serverspec'

Dir.glob(File.expand_path('../support/**/*.rb', __FILE__)).each { |f| require(f) }
#
# set :backend, :exec
