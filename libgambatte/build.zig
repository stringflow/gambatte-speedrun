const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const dynamic = b.option(bool, "dynamic", "Build the dynamic version of the library (default: false)") orelse false;

    const lib = b.addLibrary(.{
        .name = "gambatte",
        .linkage = if (dynamic) .dynamic else .static,
        .root_module = b.createModule(.{
            .target = target,
            .optimize = optimize,
            .link_libcpp = true,
        }),
    });

    lib.addIncludePath(b.path("src"));
    lib.addIncludePath(b.path("include"));
    lib.addIncludePath(b.path("../common"));
    lib.installHeadersDirectory(b.path("include"), "", .{});

    lib.addCSourceFiles(.{ .files = &c_src_files, .flags = &cflags });
    lib.addCSourceFiles(.{ .files = &cpp_src_files, .flags = &cxxflags });

    lib.root_module.addCMacro("HAVE_STDINT_H", "");
    lib.root_module.addCMacro("REVISION", rev(b.allocator));

    const zlib_dep = b.dependency("zlib", .{
        .target = target,
        .optimize = optimize,
    });
    lib.linkLibrary(zlib_dep.artifact("z"));

    b.installArtifact(lib);
}

fn rev(allocator: std.mem.Allocator) []const u8 {
    const result = std.process.Child.run(.{
        .allocator = allocator,
        .argv = &.{ "git", "rev-list", "HEAD", "--count" },
    }) catch return "-1";
    allocator.free(result.stderr);

    return result.stdout[0 .. result.stdout.len - 1];
}

const cflags = [_][]const u8{ "-Wall", "-Wextra", "-O2", "-fomit-frame-pointer" };
const cxxflags = [_][]const u8{ "-std=c++11", "-fno-exceptions", "-fno-rtti" } ++ cflags;

const c_src_files = [_][]const u8{
    "src/file/unzip/unzip.c",
    "src/file/unzip/ioapi.c",
    "src/file/file_zip.cpp",
};

const cpp_src_files = [_][]const u8{
    "src/bitmap_font.cpp",
    "src/cpu.cpp",
    "src/gambatte.cpp",
    "src/initstate.cpp",
    "src/interrupter.cpp",
    "src/interruptrequester.cpp",
    "src/loadres.cpp",
    "src/memory.cpp",
    "src/sound.cpp",
    "src/state_osd_elements.cpp",
    "src/statesaver.cpp",
    "src/tima.cpp",
    "src/video.cpp",
    "src/mem/cartridge.cpp",
    "src/mem/huc3.cpp",
    "src/mem/memptrs.cpp",
    "src/mem/pakinfo.cpp",
    "src/mem/rtc.cpp",
    "src/mem/sgb.cpp",
    "src/mem/time.cpp",
    "src/sound/channel1.cpp",
    "src/sound/channel2.cpp",
    "src/sound/channel3.cpp",
    "src/sound/channel4.cpp",
    "src/sound/duty_unit.cpp",
    "src/sound/envelope_unit.cpp",
    "src/sound/length_counter.cpp",
    "src/video/ly_counter.cpp",
    "src/video/lyc_irq.cpp",
    "src/video/next_m0_time.cpp",
    "src/video/ppu.cpp",
    "src/video/sprite_mapper.cpp",
    "src/file/file.cpp",
};
