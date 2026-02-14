import "@typespec/http";
import "@typespec/openapi";
import "@typespec/openapi3";

using TypeSpec.Http;
using TypeSpec.OpenAPI;

@service(#{
  title: "fortee API",
})
@info(#{
  version: "0.1.0",
})
namespace ForteeApi;

@route("/api/user/login")
@post
@operationId("postLogin")
op postLogin(
  @header contentType: "application/x-www-form-urlencoded",
  @body body: {
    username: string;
    password: string;
  },
): {
  @body body: {
    loggedIn: boolean;
    user?: {
      username: string;
    };
  };
};

@route("/api/user/view/{username}")
@get
@operationId("getUser")
op getUser(@path username: string): {
  @body body: {
    uuid: string;
    username: string;
    avatar_url: string;
  };
} | {
  @statusCode statusCode: 404;
};
