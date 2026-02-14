import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

import "./models.tsp";
import "./routes.tsp";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

@service(#{
  title: "Albatross internal web API",
})
@info(#{
  version: "0.3.0",
})
namespace AlbatrossApi;
