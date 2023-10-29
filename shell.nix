{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    go
    protoc-gen-go
    protoc-gen-go-grpc
    protobuf
    #go-protobuf
  ];
}

