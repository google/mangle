use super::*;

#[test]
fn test_pair_representation() {
    let mut ir = Ir::new();
    let c_1 = ir.add_inst(Inst::Number(1));
    let c_2 = ir.add_inst(Inst::Number(2));

    // Attempt to represent fn:pair(1, 2)
    // There is no Inst::Pair. We must use ApplyFn.
    let fn_pair = ir.intern_name("fn:pair");
    let pair = ir.add_inst(Inst::ApplyFn {
        function: fn_pair,
        args: vec![c_1, c_2],
    });

    match ir.get(pair) {
        Inst::ApplyFn { function, args } => {
            assert_eq!(ir.resolve_name(*function), "fn:pair");
            assert_eq!(args.len(), 2);
        }
        _ => panic!("Expected ApplyFn"),
    }
}

#[test]
fn test_struct_representation_limits() {
    let mut ir = Ir::new();
    let c_1 = ir.add_inst(Inst::Number(1));

    // Rust IR Struct requires String fields.
    // This effectively limits struct labels to strings (Names).
    let foo_name = ir.intern_name("/foo");
    let s1 = ir.add_inst(Inst::Struct {
        fields: vec![foo_name],
        values: vec![c_1],
    });

    match ir.get(s1) {
        Inst::Struct { fields, values } => {
            assert_eq!(ir.resolve_name(fields[0]), "/foo");
            assert_eq!(values[0], c_1);
        }
        _ => panic!("Expected Struct"),
    }
}

#[test]
fn test_list_and_map_representation() {
    let mut ir = Ir::new();
    let c_1 = ir.add_inst(Inst::Number(1));
    let s_foo = ir.intern_string("foo");
    let c_foo = ir.add_inst(Inst::String(s_foo));

    // List: [1, "foo"]
    let list_inst = ir.add_inst(Inst::List(vec![c_1, c_foo]));
    match ir.get(list_inst) {
        Inst::List(args) => {
            assert_eq!(args.len(), 2);
            assert_eq!(args[0], c_1);
            assert_eq!(args[1], c_foo);
        }
        _ => panic!("Expected List"),
    }

    // Map: [1: "foo"]
    let map_inst = ir.add_inst(Inst::Map {
        keys: vec![c_1],
        values: vec![c_foo],
    });
    match ir.get(map_inst) {
        Inst::Map { keys, values } => {
            assert_eq!(keys.len(), 1);
            assert_eq!(keys[0], c_1);
            assert_eq!(values[0], c_foo);
        }
        _ => panic!("Expected Map"),
    }
}

#[test]
fn test_nested_structures() {
    let mut ir = Ir::new();
    let c_1 = ir.add_inst(Inst::Number(1));

    // List of Lists: [[1]]
    let inner_list = ir.add_inst(Inst::List(vec![c_1]));
    let outer_list = ir.add_inst(Inst::List(vec![inner_list]));

    match ir.get(outer_list) {
        Inst::List(args) => {
            assert_eq!(args[0], inner_list);
        }
        _ => panic!("Expected List"),
    }

    // Map with Struct Value: [1: {foo: 1}]
    let foo_name = ir.intern_name("/foo");
    let struct_val = ir.add_inst(Inst::Struct {
        fields: vec![foo_name],
        values: vec![c_1],
    });
    let map_inst = ir.add_inst(Inst::Map {
        keys: vec![c_1],
        values: vec![struct_val],
    });

    match ir.get(map_inst) {
        Inst::Map { values, .. } => {
            assert_eq!(values[0], struct_val);
        }
        _ => panic!("Expected Map"),
    }
}
