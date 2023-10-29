{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    protobuf
    protoc-gen-go
    protoc-gen-go-grpc
    #go-protobuf
    gnumake
  ];

  shellHook = ''
    export GOENV=$(pwd)/.go/.config/go/env
    export GOMODCACHE=$(pwd)/.go/pkg/mod
    export GOPATH=$(pwd)/.go
  '';
}

