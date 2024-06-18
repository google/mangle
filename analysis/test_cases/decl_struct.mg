Decl o(O, P)
  bound[
    fn:Struct(fn:opt(/foo, fn:String())),
    fn:Struct(
        /inputs, fn:List(
            fn:Struct(
                /type, /string,
                /repeated, fn:Union(fn:Singleton(/true), fn:Singleton(/false))
            )
        ),
        /output_type, /string,
      )
  ].

