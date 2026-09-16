import Foundation
import XCTest
@testable import KumoCoreKit

final class KumoServiceManagerTests: XCTestCase {
    func testHelperCandidatesFindBundledHelperForCLIInHelpersDirectory() throws {
        let bundleURL = URL(fileURLWithPath: "/Applications/Kumo.app/Contents/Helpers", isDirectory: true)
        let workingDirectory = URL(fileURLWithPath: "/Users/test/projects", isDirectory: true)

        let candidates = KumoServiceManager.helperExecutableCandidates(
            bundleURL: bundleURL,
            executableURL: bundleURL.appendingPathComponent("kumo"),
            workingDirectory: workingDirectory,
            installedHelperURL: URL(fileURLWithPath: "/Users/test/Library/Application Support/Kumo/KumoService")
        )

        let bundledHelper = URL(fileURLWithPath: "/Applications/Kumo.app/Contents/MacOS/KumoService")
        XCTAssertTrue(candidates.contains { $0.path == bundledHelper.path })

        let bundledIndex = try XCTUnwrap(candidates.firstIndex { $0.path == bundledHelper.path })
        let workingDirectoryIndex = try XCTUnwrap(
            candidates.firstIndex { $0.path == workingDirectory.appendingPathComponent("KumoService").path }
        )
        XCTAssertLessThan(bundledIndex, workingDirectoryIndex)
    }

    func testHelperCandidatesPreferAppBundleHelperForGUIResolution() {
        let bundleURL = URL(fileURLWithPath: "/Applications/Kumo.app", isDirectory: true)

        let candidates = KumoServiceManager.helperExecutableCandidates(
            bundleURL: bundleURL,
            executableURL: bundleURL.appendingPathComponent("Contents/MacOS/KumoApp"),
            workingDirectory: URL(fileURLWithPath: "/Users/test/projects", isDirectory: true),
            installedHelperURL: URL(fileURLWithPath: "/Users/test/Library/Application Support/Kumo/KumoService")
        )

        XCTAssertEqual(candidates.first?.path, "/Applications/Kumo.app/Contents/MacOS/KumoService")
    }
}
