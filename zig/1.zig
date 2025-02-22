const std = @import("std");
const info = std.log.info;
const err = std.log.err;
const indexOf = std.mem.indexOf;
const Connection = std.net.Server.Connection;
const Thread = std.Thread;

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
            const thread = try Thread.spawn(.{}, handleConn, .{conn});
            thread.detach();
        } else |e| {
            err("error accepting connection: {}", .{e});
        }
    }
}

pub fn handleConn(conn: Connection) !void {
    info("{}> connection accepted", .{conn.address});
    defer {
        conn.stream.close();
        info("{}> connection closed", .{conn.address});
    }

    var buffer: [4096]u8 = undefined;
    while (true) {
        const bytes_received = try conn.stream.read(&buffer);
        const chunk = buffer[0..bytes_received];
        if (chunk.len == 0) {
            info("{}> done writing", .{conn.address});
            break;
        }
        _ = try conn.stream.write(chunk);
    }
}
