import { registerAccessibilitySuite } from "./accessibility-matrix.spec";
import { registerJourneySuite } from "./site-matrix.spec";
import { registerManagementSuite } from "./management-matrix.spec";
import { registerPerformanceSuite } from "./performance-matrix.spec";
import { registerResilienceSuite } from "./resilience-matrix.spec";
import { registerResponsiveSuite } from "./responsive-matrix.spec";
import { registerSecuritySuite } from "./security-matrix.spec";
import { registerVisualSuite } from "./visual-matrix.spec";

export function registerProductSuite(product: string) {
  const suite = process.env.RESOURCE_E2E_SUITE?.trim() || "all";
  if (suite === "all" || suite === "journeys") registerJourneySuite(product);
  if (suite === "all" || suite === "management")
    registerManagementSuite(product);
  if (suite === "all" || suite === "security") registerSecuritySuite(product);
  if (suite === "all" || suite === "resilience")
    registerResilienceSuite(product);
  if (suite === "all" || suite === "visual") registerVisualSuite(product);
  if (suite === "all" || suite === "visual" || suite === "responsive")
    registerResponsiveSuite(product);
  if (suite === "all" || suite === "accessibility")
    registerAccessibilitySuite(product);
  if (suite === "all" || suite === "performance")
    registerPerformanceSuite(product);
}
