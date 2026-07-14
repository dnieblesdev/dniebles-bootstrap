class Dbootstrap < Formula
  desc "Apply a declarative developer bootstrap catalog"
  homepage "https://github.com/dnieblesdev/dniebles-bootstrap"
  version "0.1.0"
  license "MIT"
  url "https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/v0.1.0/dbootstrap_v0.1.0_linux_amd64.tar.gz"
  sha256 "a8f21a55019ff09c08a124f30bffc6831c960be81cbd1496e43b26c92784d109"

  on_macos do
    disable! date: "2026-07-14", because: "dbootstrap supports Linux and WSL only; macOS is unsupported"
  end

  on_linux do
    on_arm do
      url "https://github.com/dnieblesdev/dniebles-bootstrap/releases/download/v0.1.0/dbootstrap_v0.1.0_linux_arm64.tar.gz"
      sha256 "8732f1e03ba4dc0d1a6132dd74a3291364e615aff8c52bc67727ff3f0999de6e"
    end
  end

  def install
    bin.install "dbootstrap"
    pkgshare.install "catalog/bootstrap.toml"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/dbootstrap --version")
  end
end
