"""Updater unit tests; Nix integration is exercised by the Nix workflow."""
import importlib.machinery
import importlib.util
from pathlib import Path
import unittest


class UpdaterTests(unittest.TestCase):
    def test_extract_only_vendor_derivation_hash(self):
        path = Path(__file__).with_name("update-vendor-hash")
        self.assertTrue(path.exists(), "vendor-hash updater is missing")
        loader = importlib.machinery.SourceFileLoader("updater", str(path))
        spec = importlib.util.spec_from_loader(loader.name, loader)
        assert spec is not None
        updater = importlib.util.module_from_spec(spec)
        loader.exec_module(updater)
        digest = "sha256-EMpIXQjZdkxP1lx2Zu4fWojXLtb4JlJZ+JUBn5z4yYY="
        log = "error: hash mismatch in fixed-output derivation '/nix/store/abc-ical-filter-proxy-dev-go-modules.drv':\n  specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n  got:    " + digest
        self.assertEqual(updater.extract_hash(log), digest)
        with self.assertRaises(ValueError):
            updater.extract_hash(log.replace("ical-filter-proxy-dev-go-modules", "unrelated"))
        with self.assertRaises(ValueError):
            updater.extract_hash("error: network unavailable")
        with self.assertRaises(ValueError):
            updater.extract_hash(log + "\n" + log)


    def test_update_verifies_before_writing(self):
        import subprocess
        import tempfile
        from unittest.mock import patch
        loader = importlib.machinery.SourceFileLoader(
            "updater", str(Path(__file__).with_name("update-vendor-hash")))
        spec = importlib.util.spec_from_loader(loader.name, loader)
        assert spec is not None
        updater = importlib.util.module_from_spec(spec)
        loader.exec_module(updater)
        self.assertTrue(hasattr(updater, "update"), "update operation is missing")
        digest = "sha256-EMpIXQjZdkxP1lx2Zu4fWojXLtb4JlJZ+JUBn5z4yYY="
        original = 'vendorHash = "sha256-OKOfPkssluK9uWE8BLZ9iyVpC1aLvQqzm6NyRyUYQY0=";\n'
        log = "hash mismatch in fixed-output derivation '/nix/store/abc-ical-filter-proxy-dev-go-modules.drv':\n specified: sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\n got: " + digest
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            flake = root / "flake.nix"
            flake.write_text(original)
            (root / "flake.lock").write_text("{}")
            for verified in (False, True):
                flake.write_text(original)
                responses = [subprocess.CompletedProcess([], 1, log),
                             subprocess.CompletedProcess([], 0 if verified else 1, "verification")]
                with patch.object(updater.subprocess, "run", side_effect=responses):
                    if verified:
                        updater.update(root)
                        self.assertIn(digest, flake.read_text())
                    else:
                        with self.assertRaises(RuntimeError):
                            updater.update(root)
                        self.assertEqual(flake.read_text(), original)


if __name__ == "__main__":
    unittest.main()
