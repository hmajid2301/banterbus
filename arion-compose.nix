{pkgs, ...}: let
  neat =
    pkgs.effects.runArion
    {
      name = "neat-project";
      # ignores arion-pkgs.nix even if present
      modules = [
        {}
      ];
      userSetupScript = ''
        # ...
        echo "whatever"
      '';
    }
    .prebuilt;
in {
  project.name = "banterbus";
  services = {
    otel-collector = {
      service.image = "otel/opentelemetry-collector-dev:latest";
      service.ports = [
        "4317:4317"
        "55680:55680"
      ];
    };

    jaeger = {
      service.image = "jaegertracing/all-in-one:latest";
      service.ports = [
        "16686:16686"
      ];
    };

    prometheus = {
      service.image = "prom/prometheus";
      service.ports = [
        "9090:9090"
      ];
      # service.volumes = [
      #   "${toString ./.}/prometheus.yml:/etc/prometheus/prometheus.yml"
      # ];
    };
  };
}
