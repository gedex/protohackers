const std = @import("std");
const info = std.log.info;
const err = std.log.err;
const indexOf = std.mem.indexOf;
const Connection = std.net.Server.Connection;
const Thread = std.Thread;
const Allocator = std.mem.Allocator;
const eql = std.mem.eql;
const json = std.json;
const BigInt = i256;

const Request = struct {
    method: []const u8,
    number: BigInt,
};

pub fn usage(name: []u8) void {
    info("usage: {s} <host:port>", .{name});
    std.posix.exit(1);
}

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    const allocator = gpa.allocator();
    defer _ = gpa.deinit();

    const args = try std.process.argsAlloc(allocator);
    defer std.process.argsFree(allocator, args);
    if (args.len < 2) {
        usage(args[0]);
    }

    const sep = indexOf(u8, args[1], ":").?;
    const host = args[1][0..sep];
    const port = try std.fmt.parseInt(u16, args[1][sep+1..], 10);

    const addr = try std.net.Address.resolveIp(host, port);
    var server = try addr.listen(.{ .reuse_address = true });
    info("listening on {}", .{addr});

    while (true) {
        if (server.accept()) |conn| {
            const thread = try Thread.spawn(.{}, handleConn, .{allocator, conn});
            thread.detach();
        } else |e| {
            err("error accepting connection: {}", .{e});
        }
    }
}

pub fn handleConn(allocator: Allocator, conn: Connection) !void {
    info("{}> connection accepted", .{conn.address});

    defer {
        info("{}> connection closed", .{conn.address});
        conn.stream.close();
    }

    var arr = std.ArrayList(u8).init(allocator);
    defer arr.deinit();

    while (true) {
        conn.stream.reader().streamUntilDelimiter(arr.writer(), '\n', null) catch |e| switch (e) {
            error.EndOfStream => break,
            else => unreachable,
        };
        info("{}> --> {s}", .{conn.address, arr.items});
        const request = parseRequest(allocator, arr.items) catch {
            try handleMalformedRequest(conn);
            info("{} <-- malformed", .{conn.address});
            break;
        };

        var is_prime: bool = false;
        if (request.number > 0) {
            is_prime = isNumPrime(request.number);
        }

        info("{}> <-- prime: {}", .{conn.address, is_prime});
        try handleConformingRequest(conn, is_prime);
        arr.clearRetainingCapacity();
    }
}

fn parseRequest(allocator: Allocator, message: []u8) !Request {
    var request: Request = undefined;

    const parsed = try json.parseFromSlice(
        json.Value,
        allocator,
        message,
        .{ .ignore_unknown_fields = true },
    );
    defer parsed.deinit();

    const method: ?json.Value = parsed.value.object.get("method");
    const number: ?json.Value = parsed.value.object.get("number");
    if (method == null or number == null) {
        return error.ParseError;
    }
    switch (method.?) {
        .string => {
            if (!eql(u8, method.?.string, "isPrime")) {
                return error.ParseError;
            }
            request.method = "isPrime";
        },
        else => return error.ParseError,
    }
    switch (number.?) {
        .integer => request.number = number.?.integer,
        .float => request.number = @intFromFloat(number.?.float),
        .number_string => {
            var big = try std.math.big.int.Managed.init(allocator);
            defer big.deinit();

            try big.setString(10, number.?.number_string);
            request.number = try big.to(BigInt);
        },
        else => return error.ParseError,
    }

    return request;
}

fn handleMalformedRequest(conn: Connection) !void {
    _ = try conn.stream.write("\n");
}

fn handleConformingRequest(conn: Connection, isPrime: bool) !void {
    _ = try conn.stream.writer().print("{{\"method\":\"isPrime\",\"prime\":{}}}\n", .{isPrime});
}

fn isNumPrime(n: BigInt) bool {
    if (n <= 1) return false;

    var i: isize = 2;
    while (i * i <= n) : (i += 1) {
        if (@rem(n, i) == 0) return false;
    }

    return true;
}
