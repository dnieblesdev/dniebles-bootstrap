FORMULA_PATH = File.expand_path("../Formula/dbootstrap.rb", __dir__)
RELEASE_BASE = "https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/v0.1.0"

def assert_includes(source, want)
  abort("formula must contain #{want.inspect}") unless source.include?(want)
end

def assert_before(source, first, second)
  first_index = source.index(first)
  second_index = source.index(second)
  abort("formula must contain #{first.inspect}") unless first_index
  abort("formula must contain #{second.inspect}") unless second_index
  abort("formula must place #{first.inspect} before #{second.inspect}") unless first_index < second_index
end

source = File.read(FORMULA_PATH)

assert_includes(source, "version \"0.1.0\"")
assert_includes(source, "#{RELEASE_BASE}/dbootstrap_v0.1.0_linux_amd64.tar.gz")
assert_includes(source, "sha256 \"a8f21a55019ff09c08a124f30bffc6831c960be81cbd1496e43b26c92784d109\"")
assert_includes(source, "#{RELEASE_BASE}/dbootstrap_v0.1.0_linux_arm64.tar.gz")
assert_includes(source, "sha256 \"8732f1e03ba4dc0d1a6132dd74a3291364e615aff8c52bc67727ff3f0999de6e\"")
assert_includes(source, "bin.install \"dbootstrap\"")
assert_includes(source, "pkgshare.install \"catalog/bootstrap.toml\"")
assert_includes(source, "disable! date:")
assert_includes(source, "because: \"dbootstrap supports Linux and WSL only; macOS is unsupported\"")
assert_includes(source, "test do")
assert_includes(source, "shell_output(\"#{bin}/dbootstrap --version\")")
assert_includes(source, "assert_match version.to_s")
abort("formula must not include placeholders") if source.match?(/latest|prerelease|TODO|<[^>]+>/i)

assert_before(source, "on_macos do", "on_linux do")
assert_before(source, "disable!", "on_linux do")
assert_before(source, "on_linux do", "Hardware::CPU.intel?")
assert_before(source, "Hardware::CPU.intel?", "Hardware::CPU.arm?")
assert_before(source, "Hardware::CPU.arm?", "Linux amd64 and arm64 only")
